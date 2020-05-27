package database

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/syncronizer"
	"math/big"
	"sort"
	"time"
)

type transactionChartData struct {
	transactions, failedTransactions uint64

	gasUsed, txFees *big.Int

	gasPrices map[string]int
	gasLevels map[string]int

	contractCalls, contractsDeployed uint64
}

func (b *transactionChartData) Add(tcd elem) {
	b.transactions += tcd.(*transactionChartData).transactions
	b.failedTransactions += tcd.(*transactionChartData).failedTransactions

	b.gasUsed = b.gasUsed.Add(b.gasUsed, tcd.(*transactionChartData).gasUsed)
	b.txFees = b.txFees.Add(b.txFees, tcd.(*transactionChartData).txFees)

	for k, v := range tcd.(*transactionChartData).gasPrices {
		if _, ok := b.gasPrices[k]; ok {
			b.gasPrices[k] += v
		} else {
			b.gasPrices[k] = v
		}
	}

	for k, v := range tcd.(*transactionChartData).gasLevels {
		if _, ok := b.gasLevels[k]; ok {
			b.gasLevels[k] += v
		} else {
			b.gasLevels[k] = v
		}
	}

	b.contractsDeployed += tcd.(*transactionChartData).contractsDeployed
	b.contractCalls += tcd.(*transactionChartData).contractCalls
}

func (c *Crawler) CrawlTransactions() {
	// time.Location here must be set to local
	endOfYesterday := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 23, 59, 59, 0, time.Local)

	c.logger.Warn("crawling transactions from", "from", 0, "to", endOfYesterday.Format("02/01/06"))

	cursor, err := c.backend.IterTransactions(0, endOfYesterday.Unix())
	if err != nil {
		c.logger.Error("Error creating transactions iter", "err", err)
	}

	sync := syncronizer.NewSync(20)

	result := chartData{}
	result.init()

	var transaction models.Transaction

	defer cursor.Close(context.TODO())

	for cursor.Next(context.TODO()) {

		if err := cursor.Decode(&transaction); err != nil {
			c.logger.Error("Error decoding transaction", "err", err, "transaction", transaction)
			return
		}

		c.logger.Debug("decoded transaction", "block", transaction.BlockNumber)

		currentTransaction := transaction

		sync.AddLink(func(task *syncronizer.Task) {

			var (
				gasUsed = big.NewInt(0)
				txFees  = big.NewInt(0)

				gasPrice = big.NewInt(0)
				gas      = big.NewInt(0)

				gasPrices = make(map[string]int)
				gasLevels = make(map[string]int)
			)

			mined := time.Unix(int64(currentTransaction.Timestamp), 0)
			ts := mined.Format("01/02/06")

			gasPrice = gasPrice.SetUint64(currentTransaction.GasPrice)

			gas = gas.SetUint64(currentTransaction.Gas)

			gasUsed = gasUsed.SetUint64(currentTransaction.GasUsed)

			txFees = txFees.Mul(gasUsed, gasPrice)

			gasPrices[gasPrice.String()] = 1
			gasLevels[gas.String()] = 1

			aborted := task.Link()
			if aborted {
				return
			}

			d := &transactionChartData{
				transactions: 1,
				gasUsed:      gasUsed,
				txFees:       txFees,
				gasPrices:    gasPrices,
				gasLevels:    gasLevels,
			}

			if !currentTransaction.Status && currentTransaction.BlockNumber >= 1075090 {
				d.failedTransactions = 1
			}

			if currentTransaction.ContractAddress != "" {
				d.contractsDeployed = 1
			} else if currentTransaction.Input != "0x" {
				d.contractCalls = 1
			}

			result.addElement(ts, d)
		})
	}

	aborted := sync.Finish()

	if aborted {
		if err := cursor.Err(); err != nil {
			c.logger.Error("aborted sync: error with iter", "err", err)
		} else {
			c.logger.Error("aborted sync")
		}
	}

	var (
		transactions       = make([]uint64, 0)
		failedTransactions = make([]uint64, 0)

		gasUsed = make([]string, 0)
		txFees  = make([]string, 0)

		gasPrices = make([][][]uint64, 0)
		gasLevels = make([][][]uint64, 0)

		contractCalls     = make([]uint64, 0)
		contractsDeployed = make([]uint64, 0)
	)

	dates := result.getDates()

	for _, v := range dates {
		e := result.getElement(v)
		elem := e.(*transactionChartData)

		transactions = append(transactions, elem.transactions)
		failedTransactions = append(failedTransactions, elem.failedTransactions)

		//TODO: check if overflows uint64
		gasUsed = append(gasUsed, elem.gasUsed.String())
		txFees = append(txFees, elem.txFees.String())

		prices := make([][]uint64, 0)
		for gasPrice, txns := range elem.gasPrices {

			//we discard SetString() bool return because the map keys can be set to a big.Int for sure
			gp, _ := big.NewInt(0).SetString(gasPrice, 10)

			prices = append(prices, []uint64{gp.Uint64(), uint64(txns)})
		}
		sort.Slice(prices, func(i, j int) bool {
			return prices[i][0] > prices[j][0]
		})

		gasPrices = append(gasPrices, prices)

		levels := make([][]uint64, 0)
		for gasLevel, txns := range elem.gasLevels {
			//we discard SetString() bool return because the map keys can be set to a big.Int for sure
			gl, _ := big.NewInt(0).SetString(gasLevel, 10)

			levels = append(levels, []uint64{gl.Uint64(), uint64(txns)})
		}
		sort.Slice(levels, func(i, j int) bool {
			return levels[i][0] > levels[j][0]
		})

		gasLevels = append(gasLevels, levels)

		contractCalls = append(contractCalls, elem.contractCalls)
		contractsDeployed = append(contractsDeployed, elem.contractsDeployed)

	}

	c.logger.Info("gathered chart data", "from", dates[0], "to", dates[len(dates)-1])

	err = c.backend.AddNumberChart("transactions", transactions, dates)
	if err != nil {
		c.logger.Error("error adding transactions chart", "err", err)
	}
	c.logger.Info("added transactions chart")

	err = c.backend.AddNumberChart("failedTransactions", failedTransactions, dates)
	if err != nil {
		c.logger.Error("error adding failedTransactions chart", "err", err)
	}
	c.logger.Info("added failedTransactions chart")

	err = c.backend.AddNumberStringChart("gasUsed", gasUsed, dates)
	if err != nil {
		c.logger.Error("error adding gasUsed chart", "err", err)
	}
	c.logger.Info("added gasUsed chart")

	err = c.backend.AddNumberStringChart("txFees", txFees, dates)
	if err != nil {
		c.logger.Error("error adding txFees chart", "err", err)
	}
	c.logger.Info("added txFees chart")

	err = c.backend.AddMLChart("gasPrices", gasPrices, dates)
	if err != nil {
		c.logger.Error("error adding gasPrices chart", "err", err)
	}
	c.logger.Info("added gasPrices chart")

	err = c.backend.AddMLChart("gasLevels", gasLevels, dates)
	if err != nil {
		c.logger.Error("error adding gasLevels chart", "err", err)
	}
	c.logger.Info("added gasLevels chart")

	err = c.backend.AddNumberChart("contractCalls", contractCalls, dates)
	if err != nil {
		c.logger.Error("error adding contractCalls chart", "err", err)
	}
	c.logger.Info("added contractCalls chart")

	err = c.backend.AddNumberChart("contractsDeployed", contractsDeployed, dates)
	if err != nil {
		c.logger.Error("error adding contractsDeployed chart", "err", err)
	}
	c.logger.Info("added contractsDeployed chart")

}
