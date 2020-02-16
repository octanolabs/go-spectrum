package crawler

import (
	"fmt"
	"math/big"
	stdSync "sync"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/models"

	"github.com/octanolabs/go-spectrum/crawler/syncronizer"
)

type data struct {
	avgGasPrice, txFees *big.Int
	tokentransfers      int
	stdSync.Mutex
}

func (c *Crawler) SyncLoop() {
	var currentBlock uint64

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

	sync := syncronizer.NewSync(c.cfg.MaxRoutines)

	c.state.syncing = true
	currentBlock = indexHead.Number + 1

	for ; currentBlock <= chainHead; currentBlock++ {

		sync.AddLink(func(r *syncronizer.Task) {
			block, err := c.rpc.GetBlockByHeight(currentBlock)

			if err != nil {
				log.Errorf("Error getting block: %v", err)
				c.state.syncing = false
				sync.Abort()
				return
			}

			abort := r.Link()

			if abort {
				log.Debug("Aborting routine")
				return
			}

			c.syncBlock(block, sync)
		})

	}

	abort := sync.Finish()

	if abort {
		log.Error("Aborted sync")
	}

	c.state.syncing = false
}

func (c *Crawler) syncBlock(block models.Block, sync *syncronizer.Synchronizer) {

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
	}

	pSupply = prevBlock.Supply
	pHash = prevBlock.Hash

	if pHash != block.ParentHash {
		// If pHash != to currBlock's parentHash, pHash has reorg'd
		// we remove phash from blocks collection and insert into Forkedblocks collection
		// then we abort sync so that we can sync missing blocks
		c.handleReorg(block)

		sync.Abort()
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

	sync.Log(block.Number, block.Txs, tokenTransfers, block.UncleNo, minted, supply)
}

func (c *Crawler) syncForkedBlock(b models.Block) {

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

func (c *Crawler) processTransactions(txs []models.RawTransaction, timestamp uint64) (*big.Int, *big.Int, int) {

	var twg stdSync.WaitGroup

	data := &data{
		avgGasPrice:    big.NewInt(0),
		txFees:         big.NewInt(0),
		tokentransfers: 0,
	}

	twg.Add(len(txs))

	for _, v := range txs {
		go c.processTransaction(v, timestamp, data, &twg)
	}
	twg.Wait()
	return data.avgGasPrice.Div(data.avgGasPrice, big.NewInt(int64(len(txs)))), data.txFees, data.tokentransfers
}

func (c *Crawler) processTransaction(rt models.RawTransaction, timestamp uint64, data *data, twg *stdSync.WaitGroup) {

	v := rt.Convert()

	// Create a channel

	ch := make(chan struct{}, 1)

	receipt, err := c.rpc.GetTxReceipt(v.Hash)
	if err != nil {
		log.Errorf("Error getting tx receipt: %v", err)
	}

	data.Lock()
	data.avgGasPrice.Add(data.avgGasPrice, big.NewInt(0).SetUint64(v.GasPrice))
	data.Unlock()

	gasprice := big.NewInt(0).SetUint64(v.GasPrice)

	data.Lock()
	data.txFees.Add(data.txFees, big.NewInt(0).Mul(gasprice, big.NewInt(0).SetUint64(receipt.GasUsed)))
	data.Unlock()

	v.Timestamp = timestamp
	v.GasUsed = receipt.GasUsed
	v.ContractAddress = receipt.ContractAddress
	v.Logs = receipt.Logs

	if v.IsTokenTransfer() {
		// Here we fork to a goroutine to process and insert the token transfer.
		// We use the channel to block the function util the token transfer is inserted.
		data.Lock()
		data.tokentransfers++
		data.Unlock()
		go c.processTokenTransfer(v, ch)
	}

	err = c.backend.AddTransaction(&v)
	if err != nil {
		log.Errorf("Error inserting tx into backend: %#v", err)

	}

	if v.IsTokenTransfer() {
		<-ch
	}
	twg.Done()
}

func (c *Crawler) processTokenTransfer(v models.Transaction, ch chan struct{}) {

	tktx := v.GetTokenTransfer()

	tktx.BlockNumber = v.BlockNumber
	tktx.Hash = v.Hash
	tktx.Timestamp = v.Timestamp

	err := c.backend.AddTokenTransfer(&tktx)
	if err != nil {
		log.Errorf("Error processing token transfer into backend: %v", err)
	}

	ch <- struct{}{}

}

func (c *Crawler) getPreviousBlock(blockNumber uint64) (blockCache, error) {

	// get parent block info from cache

	b := blockNumber - 1

	if cached, ok := c.blockCache.Get(b); ok {
		return cached.(blockCache), nil
	} else {
		// parent block not cached, fetch from db
		log.Warnf("block %v not found in cache, retrieving from database", b)
		latestBlock, err := c.backend.BlockByNumber(b)
		if err != nil {
			return blockCache{}, err
		}
		sprev, _ := new(big.Int).SetString(latestBlock.Supply, 10)
		return blockCache{sprev, latestBlock.Hash}, nil
	}
}

func (c *Crawler) handleReorg(b models.Block) {

	// a reorg has occured
	log.Warnf("Reorg detected at block %v", b.Number-1)

	// clear cache
	log.Warnf("Purging block cache.")
	c.blockCache.Purge()

	// sync forked Block and remove parent Block from db
	c.syncForkedBlock(b)

	log.Warnf("Forked block %v synced and removed from blocks collection.", b.Number-1)

}
