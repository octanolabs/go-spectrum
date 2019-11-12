package crawler

import (
	"math/big"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/models"
)

func (c *Crawler) SyncLoop() {
	var currentBlock uint64

	// get db head
	indexHead, err := c.backend.LatestBlock()
	if err != nil {
		log.Errorf("Error getting latest supply block: %v", err)
	}

	// get node head
	chainHead, err := c.rpc.LatestBlockNumber()
	if err != nil {
		log.Errorf("Error getting latest block number: %v", err)
	}

	syncUtility := NewSync()

	c.state.syncing = true
	c.state.reorg = false
	currentBlock = indexHead.Number + 1

	syncUtility.setInit(currentBlock)
mainloop:
	for ; currentBlock <= chainHead; currentBlock++ {
		block, err := c.rpc.GetBlockByHeight(currentBlock)
		if err != nil {
			log.Errorf("Error getting block: %v", err)
			c.state.syncing = false
			break mainloop
		}

		if c.state.reorg == true {
			// reorg has occured, reset mainloop
			syncUtility.close(currentBlock)
			c.state.reorg = false
			c.state.syncing = false

		}

		syncUtility.add(1)

		go c.Sync(block, syncUtility)

		syncUtility.wait(c.cfg.MaxRoutines)
		syncUtility.swapChannels()
	}

	syncUtility.close(currentBlock)
	c.state.syncing = false
}

func (c *Crawler) Sync(block *models.Block, syncUtility Sync) {
	syncUtility.recieve()

	// TODO: finish refactoring

	var (
		uncles  []*models.Uncle
		pSupply = new(big.Int)
		pHash   string
	)

	// get parent block info
	prevBlock, err := c.getPreviousBlock(block.Number)

	if err != nil {
		log.Errorf("Error getting previous block: %v", err)
	}

	pSupply = prevBlock.Supply
	pHash = prevBlock.Hash

	if pHash != block.ParentHash {
		c.handleReorg(block, syncUtility)
	}

	// populate uncles
	if len(block.Uncles) > 0 {
		uncles = c.GetUncles(block.Uncles, block.Number)
	}

	// calculate rewards
	blockReward, uncleRewards, minted := AccumulateRewards(block, uncles)

	// add minted to supply
	var supply = new(big.Int)
	supply.Add(pSupply, minted)

	block = &models.Block{
		Number:       block.Number,
		Hash:         block.Hash,
		Timestamp:    block.Timestamp,
		BlockReward:  blockReward.String(),
		UncleRewards: uncleRewards.String(),
		Minted:       minted.String(),
		Supply:       supply.String(),
	}

	// write block to db
	err = c.backend.AddBlock(block)
	if err != nil {
		log.Errorf("Error adding block: %v", err)
	}

	// add required block info to cache for next iteration
	c.sbCache.Add(block.Number, blockCache{Supply: supply, Hash: block.Hash})

	syncUtility.log(block.Number, minted, supply)
	syncUtility.send(block.Number + 1)
	syncUtility.done()
}

func (c *Crawler) getPreviousBlock(blockNumber uint64) (blockCache, error) {

	// get parent block info from cache

	if cached, ok := c.sbCache.Get(blockNumber - 1); ok {
		return cached.(blockCache), nil
	} else {
		// parent block not cached, fetch from db
		log.Warnf("block %v not found in cache, retrieving from database", blockNumber-1)
		latestBlock, err := c.backend.BlockByNumber(blockNumber - 1)
		if err != nil {
			return blockCache{}, err
		}
		sprev, _ := new(big.Int).SetString(latestBlock.Supply, 10)
		return blockCache{sprev, latestBlock.Hash}, nil
	}
}

func (c *Crawler) handleReorg(b *models.Block, syncUtility Sync) {

	// a reorg has occured
	log.Warnf("Reorg detected at block %v", b.Number-1)

	// clear cache
	log.Warnf("Purging block cache.")
	c.sbCache.Purge()

	// sync forked Block and remove parent Block from db
	c.syncForkedBlock(b, syncUtility)

	log.Warnf("Forked block %v synced and removed from blocks collection.", b.Number-1)

	// update state
	c.state.reorg = true
	c.state.syncing = false
	syncUtility.close(b.Number)
}

func (c *Crawler) syncForkedBlock(b *models.Block, syncUtility Sync) {

	reorgHeight := b.Number - 1

	dbBlock, err := c.backend.BlockByNumber(reorgHeight)
	if err != nil {
		log.Errorf("Error getting forked block: %v", err)
	}

	err = c.backend.AddForkedBlock(dbBlock)
	if err != nil {
		log.Errorf("Error getting reorg'd block: %v", err)
	}

	err = c.backend.PurgeBlock(reorgHeight)
	if err != nil {
		log.Errorf("Error purging reorg'd block: %v", err)
	}

	log.Warnf("HEAD - %v %v", b.Number, b.Hash)
	log.Warnf("FORKED - %v %v", dbBlock.Number, dbBlock.Hash)
}

func (c *Crawler) GetUncles(uncles []string, height uint64) []*models.Uncle {

	var u []*models.Uncle

	for k := range uncles {
		uncle, err := c.rpc.GetUncleByBlockNumberAndIndex(height, k)
		if err != nil {
			log.Errorf("Error getting uncle: %v", err)
			return u
		}
		u = append(u, uncle)
	}
	return u
}
