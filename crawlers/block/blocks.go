package block

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/models"

	"github.com/octanolabs/go-spectrum/syncronizer"
)

func (c *Crawler) RunLoop() {
	var currentBlock uint64

	if c.state.syncing {
		c.logger.Warn("Sync already in progress; quitting.")
		return
	}

	c.logChan = make(chan *logObject)

	// get db head
	indexHead, err := c.backend.LatestBlock()
	if err != nil {
		c.logger.Error("couldn't get latest block", "err", err)
	}

	c.logger.Debug("fetched block from db", "number", indexHead)

	// get node head
	chainHead, err := c.rpc.LatestBlockNumber()
	if err != nil {
		c.logger.Error("couldn't get block number", "err", err)
	}

	taskChain := syncronizer.NewSync(c.cfg.MaxRoutines)

	c.state.syncing = true
	currentBlock = indexHead.Number + 1

	syncLogger := c.logger.New("pkg", "sync", "blockNumber", strconv.FormatInt(int64(currentBlock), 10))

	start := time.Now()

	syncLogger.Debug("started sync at", "t", start)

	startLogger(c.logChan, syncLogger)

	for ; currentBlock <= chainHead; currentBlock++ {

		// capture blockNumber
		b := currentBlock

		taskChain.AddLink(func(r *syncronizer.Task) {
			block, err := c.rpc.GetBlockByHeight(b)

			if err != nil {
				syncLogger.Error("failed getting block", "err", err)
				c.state.syncing = false
				r.AbortSync()
				return
			}

			abort := r.Link()

			if abort {
				syncLogger.Debug("Aborting routine")
				return
			}

			c.syncBlock(block, r)
		})

	}

	abort := taskChain.Finish()

	if abort {
		syncLogger.Error("aborted")
	} else {
		syncLogger.Debug("terminated sync", "t", time.Since(start))
	}

	err = c.backend.UpdateStore()

	if err != nil {
		c.logger.Error("Error updating store", "err", err)
	}

	c.state.syncing = false
	close(c.logChan)
}

func (c *Crawler) syncBlock(block models.Block, task *syncronizer.Task) {

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
		c.logger.Error("couldn't get previous block", "err", err)

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
		uncles, err = c.rpc.GetUnclesInBlock(block.Uncles, block.Number)
		if err != nil {
			c.logger.Error("couldn't get uncles", "err", err)
		}
	}

	blockReward, uncleRewards, minted := c.processUncles(&block, uncles)

	// add minted to supply
	var supply = new(big.Int)
	supply.Add(pSupply, minted)

	if len(block.Transactions) > 0 {
		avgGasPrice, txFees, tokenTransfers = c.processTransactions(block.Transactions, block.Timestamp)
	}

	minted.Add(blockReward, uncleRewards)

	block.TokenTransfers = tokenTransfers
	block.AvgGasPrice = avgGasPrice.String()
	block.TxFees = txFees.String()
	block.BlockReward = blockReward.String()
	block.UncleRewards = uncleRewards.String()
	block.Minted = minted.String()
	block.Supply = supply.String()

	// write block to db
	err = c.backend.AddBlock(&block)
	if err != nil {
		c.logger.Error("couldn't add block", "err", err)
	}

	// add required block info to cache for next iteration
	c.blockCache.Add(block.Number, blockCache{Supply: supply, Hash: block.Hash})

	c.log(block.Number, block.Txs, tokenTransfers, block.UncleNo, minted, supply)
}

func (c *Crawler) syncForkedBlock(b models.Block) {

	reorgHeight := b.Number - 1

	dbBlock, err := c.backend.BlockByNumber(reorgHeight)
	if err != nil {
		c.logger.Error("couldn't get forked block", "err", err)
	}

	err = c.backend.AddForkedBlock(&dbBlock)
	if err != nil {
		c.logger.Error("couldn't add reorg'd block", "err", err)
	}

	err = c.backend.PurgeBlock(reorgHeight)
	if err != nil {
		c.logger.Error("couldn't purge reorg'd block", "err", err)
	}

	c.logger.Warn("Synced forked block", "HEAD", fmt.Sprintf("(number: %v, hash: %v)", b.Number, b.Hash), "FORKED", fmt.Sprintf("(number: %v, hash: %v)", dbBlock.Number, dbBlock.Hash))
}

type data struct {
	gasPrice, txFees *big.Int
	tokenTransfers   int
}

func (c *Crawler) processUncles(block *models.Block, uncles []models.Uncle) (*big.Int, *big.Int, *big.Int) {

	var (
		uRewards = new(big.Int)
	)

	blockReward, uncleRewards, minted := AccumulateRewards(block, uncles)

	for idx, uncle := range uncles {

		uncle.BlockNumber = block.Number
		uncle.Position = uint64(idx)
		uncle.Reward = uncleRewards[idx].String()

		uRewards.Add(uRewards, uncleRewards[idx])

		err := c.backend.AddUncle(&uncle)

		if err != nil {
			c.logger.Error("couldn't add uncle", "uncle", uncle)
		}
	}

	return blockReward, uRewards, minted

}

func (c *Crawler) processTransactions(txs []models.RawTransaction, timestamp uint64) (avgGasPrice, txFees *big.Int, tokenTransfers int) {

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
				c.logger.Error("couldn't get tx receipt", "err", err)
			}
			closed := t.Link()

			if closed {
				return
			}

			c.processTransaction(&tx, receipt, data)
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

				c.processTokenTransfer(transfer, &tx)
			})
		}

	}

	txSync.Finish()

	return data.gasPrice.Div(data.gasPrice, big.NewInt(int64(len(txs)))), data.txFees, data.tokenTransfers
}

func (c *Crawler) processTransaction(tx *models.Transaction, receipt models.TxReceipt, data *data) {

	txGasPrice := big.NewInt(0).SetUint64(tx.GasPrice)

	data.gasPrice.Add(data.gasPrice, txGasPrice)

	txFees := big.NewInt(0).Mul(txGasPrice, big.NewInt(0).SetUint64(receipt.GasUsed))

	data.txFees.Add(data.txFees, txFees)

	tx.GasUsed = receipt.GasUsed
	tx.ContractAddress = receipt.ContractAddress
	tx.Logs = receipt.Logs
	tx.Status = receipt.Status

	c.logger.Debug("tx to insert", "tx", fmt.Sprintf("%+v", tx), "receipt", fmt.Sprintf("%+v", receipt))

	err := c.backend.AddTransaction(tx)
	if err != nil {
		c.logger.Error("couldn't insert tx into backend", "err", err)
	}

}

func (c *Crawler) processTokenTransfer(transfer *models.TokenTransfer, tx *models.Transaction) {

	// Setting status here as we need to wait for the tx in the previous link to be processed
	transfer.Status = tx.Status

	err := c.backend.AddTokenTransfer(transfer)
	if err != nil {
		log.Errorf("couldn't insert token transfer into backend", "err", err)
	}

}

func (c *Crawler) getPreviousBlock(blockNumber uint64) (blockCache, error) {

	// get parent block info from cache

	b := blockNumber - 1

	if cached, ok := c.blockCache.Get(b); ok {
		return cached.(blockCache), nil
	} else {
		// parent block not cached, fetch from db
		c.logger.Warn("block not found in cache, retrieving from database", "number", b)

		latestBlock, err := c.backend.BlockByNumber(b)
		if err != nil {
			return blockCache{}, errors.New("block " + strconv.FormatInt(int64(b), 10) + " not found in database")
		}
		sprev, _ := new(big.Int).SetString(latestBlock.Supply, 10)
		return blockCache{sprev, latestBlock.Hash}, nil
	}
}

func (c *Crawler) handleReorg(b models.Block) {

	// a reorg has occured
	c.logger.Warn("reorg detected", "height", b.Number-1)

	// clear cache
	c.logger.Warn("Purging block cache.")
	c.blockCache.Purge()

	// sync forked Block and remove parent Block from db
	c.syncForkedBlock(b)

	c.logger.Warn("synced forked block and purged reorg from blocks collection", "height", b.Number-1)

}

func (c *Crawler) log(blockNo uint64, txns, transfers, uncles int, minted *big.Int, supply *big.Int) {
	c.logChan <- &logObject{
		blockNo:        blockNo,
		txns:           txns,
		tokentransfers: transfers,
		uncleNo:        uncles,
		minted:         minted,
		supply:         supply,
	}
}
