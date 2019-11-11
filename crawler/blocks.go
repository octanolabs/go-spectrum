package crawler

import (
	"math/big"

	log "github.com/sirupsen/logrus"

	"github.com/octanolabs/go-spectrum/models"
)

func (c *Crawler) SyncLoop() {
	var currentBlock uint64

	// get db head
	indexHead, err := c.backend.LatestSupplyBlock()
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

	var (
		uncles  []*models.Uncle
		pSupply = new(big.Int)
		pHash   string
	)

	// populate uncles
	if len(block.Uncles) > 0 {
		uncles = c.GetUncles(block.Uncles, block.Number)
	}

	// calculate rewards
	blockReward, uncleRewards, minted := AccumulateRewards(block, uncles)

	// get parent block info from cache
	if cached, ok := c.sbCache.Get(block.Number - 1); ok {
		pSupply = cached.(sbCache).Supply
		pHash = cached.(sbCache).Hash
	} else {
		// parent block not cached, fetch from db
		log.Warnf("block %v not found in cache, retrieving from database", block.Number-1)
		lsb, err := c.backend.SupplyBlockByNumber(block.Number - 1)
		if err != nil {
			log.Errorf("Error getting latest supply block: %v", err)
			syncUtility.send(block.Number)
			syncUtility.done()
		} else {
			sprev, _ := new(big.Int).SetString(lsb.Supply, 10)
			pSupply = sprev
			pHash = lsb.Hash
		}
	}

	// check parent hash incase a reorg has occured.
	if pHash != block.ParentHash {
		// a reorg has occured
		log.Warnf("Reorg detected at block %v", block.Number-1)
		// clear cache
		log.Println("Purging block cache.")
		c.sbCache.Purge()
		// remove parent Block from db
		err := c.backend.RemoveSupplyBlock(block.Number - 1)
		if err != nil {
			log.Errorf("Error removing supply block: %v", err)
			syncUtility.done()
		} else {
			log.Printf("Block %v removed from db.", block.Number-1)
			// update state
			c.state.reorg = true
			c.state.syncing = false
			syncUtility.close(block.Number)
			//syncUtility.done()
		}
	} else {
		// add minted to supply
		var supply = new(big.Int)
		supply.Add(pSupply, minted)

		Block := models.Block{
			Number:       block.Number,
			Hash:         block.Hash,
			Timestamp:    block.Timestamp,
			BlockReward:  blockReward.String(),
			UncleRewards: uncleRewards.String(),
			Minted:       minted.String(),
			Supply:       supply.String(),
		}

		// write block to db
		err := c.backend.AddSupplyBlock(Block)
		if err != nil {
			log.Errorf("Error adding block: %v", err)
		}

		// add required block info to cache for next iteration
		c.sbCache.Add(block.Number, sbCache{Supply: supply, Hash: block.Hash})

		syncUtility.log(block.Number, minted, supply)
		syncUtility.send(block.Number + 1)
		syncUtility.done()
	}
}

func (c *Crawler) GetUncles(uncles []string, height uint64) []*models.Uncle {

	var u []*models.Uncle

	for k, _ := range uncles {
		uncle, err := c.rpc.GetUncleByBlockNumberAndIndex(height, k)
		if err != nil {
			log.Errorf("Error getting uncle: %v", err)
			return u
		}
		u = append(u, uncle)
	}
	return u
}
