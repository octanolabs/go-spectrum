package database

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/syncronizer"
	"math/big"
	"time"
)

type blockChartData struct {
	avgGasPrice, gasLimit, difficulty, blockTime, blocks, transactions *big.Int
	miners                                                             map[string]int
}

func (b *blockChartData) Add(bcd interface{}) {
	b.avgGasPrice.Add(b.avgGasPrice, bcd.(*blockChartData).avgGasPrice)
	b.gasLimit.Add(b.gasLimit, bcd.(*blockChartData).gasLimit)
	b.difficulty.Add(b.difficulty, bcd.(*blockChartData).difficulty)
	b.blockTime.Add(b.blockTime, bcd.(*blockChartData).blockTime)
	b.transactions.Add(b.blockTime, bcd.(*blockChartData).blockTime)
	b.blocks.Add(b.blocks, bcd.(*blockChartData).blocks)

	for k, v := range bcd.(*blockChartData).miners {
		if _, ok := b.miners[k]; ok {
			b.miners[k] += v
		} else {
			b.miners[k] = v
		}
	}
}

func (b *blockChartData) weigh() {
	b.avgGasPrice = b.avgGasPrice.Div(b.avgGasPrice, b.blocks)
	b.gasLimit = b.gasLimit.Div(b.gasLimit, b.blocks)
	b.difficulty = b.difficulty.Div(b.difficulty, b.blocks)
	b.blockTime = b.blockTime.Div(b.blockTime, b.blocks)
}

func (c *Crawler) CrawlBlocks() {

	var (
		avgGasPrice  = make([]uint64, 0)
		gasLimit     = make([]uint64, 0)
		difficulty   = make([]uint64, 0)
		blockTime    = make([]uint64, 0)
		blocks       = make([]uint64, 0)
		transactions = make([]uint64, 0)

		miners = make(map[string][]int, 0)
	)

	cursor, err := c.backend.IterBlocks()
	if err != nil {
		c.logger.Error("Error creating block iter", "err", err)
	}

	sync := syncronizer.NewSync(20)

	result := chartData{}
	result.init()

	var (
		block models.Block
		stamp uint64
	)

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {

		if err := cursor.Decode(&block); err != nil {
			c.logger.Error("Error decoding block", "err", err, "block", block)
			return
		}

		c.logger.Debug("decoded block", "number", block.Number)

		currentBlock := block
		prevStamp := stamp

		sync.AddLink(func(task *syncronizer.Task) {

			var (
				bTime  = big.NewInt(0)
				avgGas = big.NewInt(0)
				diff   = big.NewInt(0)
			)

			ts := time.Unix(int64(currentBlock.Timestamp), 0).Format("02/01/06")

			_, ok := avgGas.SetString(currentBlock.AvgGasPrice, 10)
			if !ok {
				c.logger.Error("error setting avggas", "height", currentBlock.Number, "avgGas", currentBlock.AvgGasPrice)
				return
			}

			_, ok = diff.SetString(currentBlock.Difficulty, 10)
			if !ok {
				c.logger.Error("error setting diff", "height", currentBlock.Number, "diff", currentBlock.Difficulty)
				return
			}

			if prevStamp != 0 {
				bTime = new(big.Int).SetUint64(currentBlock.Timestamp - prevStamp)
			}

			miners := make(map[string]int)
			miners[currentBlock.Miner] = 1

			aborted := task.Link()
			if aborted {
				return
			}

			d := &blockChartData{
				avgGasPrice:  avgGas,
				gasLimit:     new(big.Int).SetUint64(block.GasLimit),
				difficulty:   diff,
				blockTime:    bTime,
				blocks:       new(big.Int).SetInt64(1),
				transactions: new(big.Int).SetInt64(int64(block.Txs)),
				miners:       miners,
			}

			result.addElement(ts, d)
		})

		stamp = block.Timestamp

	}

	aborted := sync.Finish()

	if aborted {
		if err := cursor.Err(); err != nil {
			c.logger.Error("aborted sync: error with iter", "err", err)
		} else {
			c.logger.Error("aborted sync")
		}
	}

	dates := result.getDates()

	for _, v := range dates {
		e := result.getElement(v)
		elem := e.(*blockChartData)

		elem.weigh()

		avgGasPrice = append(avgGasPrice, elem.avgGasPrice.Uint64())
		gasLimit = append(gasLimit, elem.gasLimit.Uint64())
		difficulty = append(difficulty, elem.difficulty.Uint64())
		blockTime = append(blockTime, elem.blockTime.Uint64())
		blocks = append(blocks, elem.blocks.Uint64())
		transactions = append(transactions, elem.transactions.Uint64())

		for k, v := range elem.miners {
			if _, ok := miners[k]; ok {
				miners[k] = append(miners[k], v)
			} else {
				miners[k] = []int{v}
			}
		}

	}

	c.logger.Info("gathered chart data")

	c.logger.Info("added chart: gasPrice", "n", len(avgGasPrice))
	err = c.backend.AddChart("avgGasPrice", avgGasPrice, dates)
	if err != nil {
		c.logger.Error("error adding avgGasPrice chart", "err", err)
	}

	c.logger.Info("added chart: difficulty", "n", len(difficulty))
	err = c.backend.AddChart("difficulty", difficulty, dates)
	if err != nil {
		c.logger.Error("error adding difficulty chart", "err", err)
	}

	c.logger.Info("added chart: blocks", "n", len(blocks))
	err = c.backend.AddChart("blocks", blocks, dates)
	if err != nil {
		c.logger.Error("error adding blocks chart", "err", err)
	}

	c.logger.Info("added chart: gasLimit", "n", len(gasLimit))
	err = c.backend.AddChart("gasLimit", gasLimit, dates)
	if err != nil {
		c.logger.Error("error adding chart ", "err", err)
	}

	c.logger.Info("added chart: blockTime", "n", len(blockTime))
	err = c.backend.AddChart("blockTime", blockTime, dates)
	if err != nil {
		c.logger.Error("error adding gasLimit chart", "err", err)
	}

	c.logger.Info("added chart: transactions", "n", len(transactions))
	err = c.backend.AddChart("transactions", transactions, dates)
	if err != nil {
		c.logger.Error("error adding blockTime chart", "err", err)
	}

	list := make([]string, 0)

	for k, v := range miners {
		list = append(list, k)

		n := "miner_" + k

		err := c.backend.AddChart(n, v, dates)
		if err != nil {
			c.logger.Error("error adding miner chart", "miner", k, "err", err)
		}
	}

	c.logger.Info("added charts for miners", "n", len(list))
	err = c.backend.AddChart("miners", list, dates)
	if err != nil {
		c.logger.Error("error adding miners chart", "err", err)
	}
}
