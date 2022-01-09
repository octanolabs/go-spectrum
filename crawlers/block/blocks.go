package block

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"time"

	"github.com/ubiq/go-ubiq/v6/log"

	"github.com/octanolabs/go-spectrum/models"

	"github.com/octanolabs/go-spectrum/syncronizer"
)

func (c *Crawler) RunLoop() {

	c.logChan = make(chan *logObject)

	c.crawBlocks()

	close(c.logChan)

	var startBlock = c.cfg.Tracing.StartBlock

	t, err := c.backend.LatestTxTrace()
	if err != nil {
		c.logger.Debug("error: couldn't get latest tx trace", " err", err)
	} else {
		startBlock = uint64(t.OriginBlockNo + 1)
	}

	//c.logger.Info("starting tx tracing", "batchSize", c.cfg.Tracing.BatchSize, "traceFrom", startBlock)

	c.crawlTxTraces(startBlock)

	err = c.backend.UpdateStore()

	if err != nil {
		c.logger.Error("Error updating store", "err", err)
	}

	c.state.syncing = false
}

func (c *Crawler) crawlTxTraces(startBlock uint64) {

	if s, err := c.backend.Status(); err == nil && s.LatestBlock.Number < c.cfg.Tracing.StartBlock {
		c.logger.Error("skipping cycle, didn't sync past startBlock yet")
	} else {

		latestBlock, err := c.backend.LatestBlock()
		if err != nil {
			c.logger.Error("couldn't get latest block", "err", err)
			return
		}

		sync := syncronizer.NewSync(25)

		for b := startBlock; b < latestBlock.Number; b += uint64(c.cfg.Tracing.BatchSize) + 1 {
			t := time.Now()

			c.logger.Debug("syncing tx traces tx traces", "t", t.String(), "b", b)

			hashes, blockNos, err := c.backend.LatestTxHashes(c.cfg.Tracing.BatchSize, b)
			if err != nil {
				c.logger.Error("error getting trace data", "err", err)
			}

			c.logger.Debug("latest", "hashes", hashes, "bns", blockNos)

			if len(hashes) == 0 {
				c.logger.Warn("no traces to sync")
				break
			}

			c.syncTxTraces(sync, hashes, blockNos)
			c.logger.Info("synced tx traces", "head", blockNos[len(blockNos)-1], "count", len(hashes), "took", time.Since(t).String())

		}

		if aborted := sync.Finish(); aborted {
			log.Error("aborted sync")
		}
	}
}

func (c *Crawler) syncTxTraces(sync *syncronizer.Synchronizer, hashes []string, blockNos []int64) {

	for idx, txh := range hashes {

		hash := txh
		bn := blockNos[idx]

		//TODO: multiple errors causing simultaneous aborts make synchronizer hang
		// figure out why (to reproduce remove timout from trace rpc call)
		sync.AddLink(func(task *syncronizer.Task) {

			itx, err := c.rpc.TraceTransaction(hash)
			if err != nil {
				c.logger.Error("couldn't get internal tx", "hash", hash, "bn", bn, "err", err)
				task.AbortSync()
				return
			}

			if closed := task.Link(); closed {
				return
			}

			trace := &models.TxTrace{
				OriginTxHash:  hash,
				OriginBlockNo: bn,
				Trace:         itx,
			}

			err = c.backend.AddTxTrace(trace)
			if err != nil {
				log.Error("couldn't add trace to db", "err", err)
				task.AbortSync()
				return
			}
		})
	}
	return
}

func (c *Crawler) crawBlocks() {
	var currentBlock uint64

	if c.state.syncing {
		c.logger.Warn("Sync already in progress; quitting.")
		return
	}

	c.state.syncing = true

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

	currentBlock = indexHead.Number + 1

	syncLogger := c.logger.New("pkg", "sync", "blockNumber", strconv.FormatInt(int64(currentBlock), 10))
	startLogger(c.logChan, syncLogger)

	start := time.Now()

	syncLogger.Debug("started sync at", "t", start)

	taskChain := syncronizer.NewSync(c.cfg.MaxRoutines)
	//TODO: look into: if GetBlockByHeight() fails, the taskchain stops but the loop keeps going until chainHead is reached
	// maybe introduce a new metod on syncronizer like DidAbort that can be used in loop condition to quit
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
}

func (c *Crawler) syncBlock(block models.Block, task *syncronizer.Task) {

	var (
		uncles                                           = make([]models.Uncle, 0)
		avgGasPrice, txFees                              = new(big.Int), new(big.Int)
		pSupply                                          = new(big.Int)
		pTotalBurned                                     = new(big.Int)
		pHash                                            string
		tokenTransfers, contractsDeployed, contractCalls int
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
	pTotalBurned = prevBlock.TotalBurned

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

	// add burned to totalBurned
	var totalBurned = new(big.Int)
	burnedUint64, _ := new(big.Int).SetString(block.Burned, 10)
	totalBurned.Add(pTotalBurned, burnedUint64)

	// remove burned from supply
	supply.Sub(supply, burnedUint64)

	if len(block.Transactions) > 0 {
		avgGasPrice, txFees, tokenTransfers, contractsDeployed, contractCalls = c.processTransactions(block.Transactions, block.Timestamp, block.BaseFeePerGas)
	}

	minted.Add(blockReward, uncleRewards)

	block.TokenTransfers = tokenTransfers
	block.AvgGasPrice = avgGasPrice.String()
	block.TxFees = txFees.String()
	block.BlockReward = blockReward.String()
	block.UncleRewards = uncleRewards.String()
	block.Minted = minted.String()
	block.Supply = supply.String()
	block.TotalBurned = totalBurned.String()
	// write block to db
	err = c.backend.AddBlock(&block)
	if err != nil {
		c.logger.Error("couldn't add block", "err", err)
	}

	// add required block info to cache for next iteration
	c.blockCache.Add(block.Number, blockCache{Supply: supply, Hash: block.Hash, TotalBurned: totalBurned})

	c.log(block.Number, block.Txs, tokenTransfers, contractsDeployed, contractCalls, block.UncleNo, minted, supply)
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
	gasPrice, txFees                                 *big.Int
	tokenTransfers, contractCalls, contractsDeployed int
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

func (c *Crawler) processTransactions(txs []models.RawTransaction, timestamp uint64, baseFeePerGas string) (avgGasPrice, txFees *big.Int, tokenTransfers, contractsDeployed, contractCalls int) {

	data := &data{
		gasPrice:          big.NewInt(0),
		txFees:            big.NewInt(0),
		tokenTransfers:    0,
		contractCalls:     0,
		contractsDeployed: 0,
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

			c.processTransaction(&tx, receipt, data, baseFeePerGas)
		})

		//txSync.AddLink(func(t *syncronizer.Task) {
		//
		//	txTrace, err := c.rpc.TraceTransaction(tx.Hash)
		//	if err != nil {
		//		c.logger.Error("couldn't get tx receipt", "err", err)
		//	}
		//	closed := t.Link()
		//	if closed {
		//		return
		//	}
		//
		//	err = c.backend.AddTxTrace(txTrace)
		//	if err != nil {
		//		c.logger.Error("couldn't store tx trace", "err", err)
		//	}
		//})

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

	return data.gasPrice.Div(data.gasPrice, big.NewInt(int64(len(txs)))), data.txFees, data.tokenTransfers, data.contractsDeployed, data.contractCalls
}

func (c *Crawler) processTransaction(tx *models.Transaction, receipt models.TxReceipt, data *data, baseFeePerGas string) {

	txGasPrice := big.NewInt(0).SetUint64(tx.GasPrice)

	data.gasPrice.Add(data.gasPrice, txGasPrice)

	txFees := big.NewInt(0).Mul(txGasPrice, big.NewInt(0).SetUint64(receipt.GasUsed))

	data.txFees.Add(data.txFees, txFees)

	tx.GasUsed = receipt.GasUsed
	tx.ContractAddress = receipt.ContractAddress
	tx.Logs = receipt.Logs
	tx.Status = receipt.Status
	tx.BaseFeePerGas = baseFeePerGas

	err := c.backend.AddTransaction(tx)
	if err != nil {
		c.logger.Error("couldn't insert tx into backend", "err", err)
	}

	if tx.IsContractDeployTxn() {
		data.contractsDeployed++

		err := c.backend.AddDeployedContract(tx)
		if err != nil {
			c.logger.Error("couldn't insert deployed contract into backend", "err", err)
		}
	}

	if tx.IsContractCall() {
		data.contractCalls++

		err := c.backend.AddContractCall(tx)
		if err != nil {
			c.logger.Error("couldn't insert contract call into backend", "err", err)
		}
	}

}

func (c *Crawler) processTokenTransfer(transfer *models.TokenTransfer, tx *models.Transaction) {

	// Setting status here as we need to wait for the tx in the previous link to be processed
	transfer.Status = tx.Status

	err := c.backend.AddTokenTransfer(transfer)
	if err != nil {
		c.logger.Error("couldn't insert token transfer into backend", "err", err)
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
		bprev, _ := new(big.Int).SetString(latestBlock.TotalBurned, 10)
		return blockCache{sprev, latestBlock.Hash, bprev}, nil
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

func (c *Crawler) log(blockNo uint64, txns, transfers, contractsDeployed, contractCalls, uncles int, minted *big.Int, supply *big.Int) {
	c.logChan <- &logObject{
		blockNo:           blockNo,
		txns:              txns,
		tokentransfers:    transfers,
		contractCalls:     contractCalls,
		contractsDeployed: contractsDeployed,
		uncleNo:           uncles,
		minted:            minted,
		supply:            supply,
	}
}
