package crawler

import (
	"errors"
	"fmt"
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/models"

	"github.com/octanolabs/go-spectrum/syncronizer"
)

func (c *BlockCrawler) SyncLoop() {
	var currentBlock uint64

	c.logChan = make(chan *logObject)
	startLogger(c.logChan)

	start := time.Now()

	// get db head
	indexHead, err := c.backend.LatestBlock()
	if err != nil {
		log.Fatalf("Error getting latest block: %v", err)
	}

	log.Debugf("Db block: %+v", indexHead)

	// get node head
	chainHead, err := c.rpc.LatestBlockNumber()
	if err != nil {
		log.Errorf("Error getting block number: %v", err)
	}

	taskChain := syncronizer.NewSync(c.cfg.MaxRoutines)

	c.state.syncing = true
	currentBlock = indexHead.Number + 1

	for ; currentBlock <= chainHead; currentBlock++ {

		// capture blockNumber
		b := currentBlock

		taskChain.AddLink(func(r *syncronizer.Task) {
			block, err := c.rpc.GetBlockByHeight(b)

			if err != nil {
				log.Errorf("Error getting block: %v", err)
				c.state.syncing = false
				r.AbortSync()
				return
			}

			abort := r.Link()

			if abort {
				log.Debug("Aborting routine")
				return
			}

			c.syncBlock(block, r)
		})

	}

	abort := taskChain.Finish()

	if abort {
		log.Error("aborted sync")
	} else {
		log.Debugf("Terminated sync in %v", time.Since(start))
	}

	c.state.syncing = false
	close(c.logChan)
}

func (c *BlockCrawler) syncBlock(block models.Block, task *syncronizer.Task) {

	var (
		uncles              = make([]models.Uncle, 0)
		avgGasPrice, txFees = new(big.Int), new(big.Int)
		pSupply             = new(big.Int)
		pHash               string
		tokenTransfers      int
	)

	// get parent block info
	prevBlock, err := c.getPreviousBlock(block.Number)

	if err != nil {
		log.Errorf("Error getting previous block: %v", err)

		task.AbortSync()
		return
	}

	pSupply = prevBlock.Supply
	pHash = prevBlock.Hash

	if pHash != block.ParentHash {
		// If pHash != to currBlock's parentHash, pHash has reorg'd
		// we remove phash from blocks collection and insert into Forkedblocks collection
		// then we abort sync so that we can sync missing blocks
		c.handleReorg(block)

		task.AbortSync()
		return
	}

	// populate uncles
	if len(block.Uncles) > 0 {
		uncles = c.rpc.GetUnclesInBlock(block.Uncles, block.Number)
	}

	// calculate rewards
	blockReward, uncleRewards, minted := AccumulateRewards(&block, uncles)

	// add minted to supply
	var supply = new(big.Int)
	supply.Add(pSupply, minted)

	if len(block.Transactions) > 0 {
		avgGasPrice, txFees, tokenTransfers = c.processTransactions(block.Transactions, block.Timestamp)
	}

	minted.Add(blockReward, uncleRewards)

	block.AvgGasPrice = avgGasPrice.String()
	block.TxFees = txFees.String()
	block.BlockReward = blockReward.String()
	block.UncleRewards = uncleRewards.String()
	block.Minted = minted.String()
	block.Supply = supply.String()

	// write block to db
	err = c.backend.AddBlock(&block)
	if err != nil {
		log.Errorf("Error adding block: %v", err)
	}

	// add required block info to cache for next iteration
	c.blockCache.Add(block.Number, blockCache{Supply: supply, Hash: block.Hash})

	c.log(block.Number, block.Txs, tokenTransfers, block.UncleNo, minted, supply)
}

func (c *BlockCrawler) syncForkedBlock(b models.Block) {

	reorgHeight := b.Number - 1

	dbBlock, err := c.backend.BlockByNumber(reorgHeight)
	if err != nil {
		log.Errorf("Error getting forked block: %v", err)
	}

	err = c.backend.AddForkedBlock(&dbBlock)
	if err != nil {
		log.Errorf("Error adding reorg'd block: %v", err)
	}

	err = c.backend.PurgeBlock(reorgHeight)
	if err != nil {
		log.Errorf("Error purging reorg'd block: %v", err)
	}

	log.WithFields(log.Fields{
		"HEAD":   fmt.Sprint(b.Number, b.Hash),
		"FORKED": fmt.Sprint(dbBlock.Number, dbBlock.Hash),
	}).Warn("Synced forked block")
}

type data struct {
	gasPrice, txFees *big.Int
	tokenTransfers   int
}

func (c *BlockCrawler) processTransactions(txs []models.RawTransaction, timestamp uint64) (avgGasPrice, txFees *big.Int, tokenTransfers int) {

	data := &data{
		gasPrice:       big.NewInt(0),
		txFees:         big.NewInt(0),
		tokenTransfers: 0,
	}

	// maxRoutines equal to 2 times the number of txs to account for possible token transfers
	txSync := syncronizer.NewSync(len(txs) * 2)

	for _, val := range txs {

		// Capture value of rawTx
		rt := val

		tx := rt.Convert()

		// Set timestamp here, if it's a token transfer the field needs to be present
		tx.Timestamp = timestamp

		txSync.AddLink(func(t *syncronizer.Task) {

			receipt, err := c.rpc.GetTxReceipt(tx.Hash)
			if err != nil {
				log.Errorf("Error getting tx receipt: %v", err)
			}
			closed := t.Link()

			if closed {
				return
			}

			c.processTransaction(tx, receipt, data)
		})

		// If tx is a token transfer we add another link right after

		if tx.IsTokenTransfer() {

			data.tokenTransfers++
			txSync.AddLink(func(task *syncronizer.Task) {

				transfer := tx.GetTokenTransfer()

				closed := task.Link()

				if closed {
					return
				}

				c.processTokenTransfer(transfer)
			})
		}

	}

	txSync.Finish()

	return data.gasPrice.Div(data.gasPrice, big.NewInt(int64(len(txs)))), data.txFees, data.tokenTransfers
}

func (c *BlockCrawler) processTransaction(tx models.Transaction, receipt models.TxReceipt, data *data) {

	txGasPrice := big.NewInt(0).SetUint64(tx.GasPrice)

	data.gasPrice.Add(data.gasPrice, txGasPrice)

	txFees := big.NewInt(0).Mul(txGasPrice, big.NewInt(0).SetUint64(receipt.GasUsed))

	data.txFees.Add(data.txFees, txFees)

	tx.GasUsed = receipt.GasUsed
	tx.ContractAddress = receipt.ContractAddress
	tx.Logs = receipt.Logs

	err := c.backend.AddTransaction(&tx)
	if err != nil {
		log.Errorf("Error inserting tx into backend: %#v", err)
	}

}

func (c *BlockCrawler) processTokenTransfer(transfer *models.TokenTransfer) {

	err := c.backend.AddTokenTransfer(transfer)
	if err != nil {
		log.Errorf("Error processing token transfer into backend: %v", err)
	}

}

func (c *BlockCrawler) getPreviousBlock(blockNumber uint64) (blockCache, error) {

	// get parent block info from cache

	b := blockNumber - 1

	if cached, ok := c.blockCache.Get(b); ok {
		return cached.(blockCache), nil
	} else {
		// parent block not cached, fetch from db
		log.Warnf("block %v not found in cache (%v), retrieving from database", b, c.blockCache)

		latestBlock, err := c.backend.BlockByNumber(b)
		if err != nil {
			return blockCache{}, errors.New("block %v not found in database")
		}
		sprev, _ := new(big.Int).SetString(latestBlock.Supply, 10)
		return blockCache{sprev, latestBlock.Hash}, nil
	}
}

func (c *BlockCrawler) handleReorg(b models.Block) {

	// a reorg has occured
	log.Warnf("Reorg detected at block %v", b.Number-1)

	// clear cache
	log.Warnf("Purging block cache.")
	c.blockCache.Purge()

	// sync forked Block and remove parent Block from db
	c.syncForkedBlock(b)

	log.Warnf("Forked block %v synced and removed from blocks collection.", b.Number-1)

}

func (c *BlockCrawler) log(blockNo uint64, txns, transfers, uncles int, minted *big.Int, supply *big.Int) {
	c.logChan <- &logObject{
		blockNo:        blockNo,
		txns:           txns,
		tokentransfers: transfers,
		uncleNo:        uncles,
		minted:         minted,
		supply:         supply,
	}
}
