package database

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/syncronizer"
	"math/big"
	"time"
)

type transactionChartData struct {
	transactions, failedTransactions uint64

	gasUsed, txFees *big.Int

	gasPriceLevels map[string]uint
	gasUsedLevels  map[string]uint
	gasLevels      map[string]uint

	transactedValue *big.Int

	contractCalls, contractsDeployed uint64
}

func (b *transactionChartData) Add(tcd elem) {
	b.transactions += tcd.(*transactionChartData).transactions
	b.failedTransactions += tcd.(*transactionChartData).failedTransactions

	b.gasUsed = b.gasUsed.Add(b.gasUsed, tcd.(*transactionChartData).gasUsed)
	b.txFees = b.txFees.Add(b.txFees, tcd.(*transactionChartData).txFees)

	b.transactedValue = b.transactedValue.Add(b.transactedValue, tcd.(*transactionChartData).transactedValue)

	for k, v := range tcd.(*transactionChartData).gasPriceLevels {
		if _, ok := b.gasPriceLevels[k]; ok {
			b.gasPriceLevels[k] += v
		} else {
			b.gasPriceLevels[k] = v
		}
	}

	for k, v := range tcd.(*transactionChartData).gasUsedLevels {
		if _, ok := b.gasUsedLevels[k]; ok {
			b.gasUsedLevels[k] += v
		} else {
			b.gasUsedLevels[k] = v
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

				transactedValue = big.NewInt(0)

				gasPriceLevels = make(map[string]uint)
				gasUsedLevels  = make(map[string]uint)
				gasLevels      = make(map[string]uint)
			)

			mined := time.Unix(int64(currentTransaction.Timestamp), 0)
			ts := mined.Format("01/02/06")

			gasPrice = gasPrice.SetUint64(currentTransaction.GasPrice)

			gas = gas.SetUint64(currentTransaction.Gas)

			gasUsed = gasUsed.SetUint64(currentTransaction.GasUsed)

			transactedValue, _ = transactedValue.SetString(currentTransaction.Value, 10)

			txFees = txFees.Mul(gasUsed, gasPrice)

			gasPriceLevels[gasPrice.String()] = 1
			gasUsedLevels[gasUsed.String()] = 1
			gasLevels[gas.String()] = 1

			aborted := task.Link()
			if aborted {
				return
			}

			d := &transactionChartData{
				transactions:    1,
				gasUsed:         gasUsed,
				txFees:          txFees,
				gasPriceLevels:  gasPriceLevels,
				gasUsedLevels:   gasUsedLevels,
				gasLevels:       gasLevels,
				transactedValue: transactedValue,
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

		gasPriceLevels = make(map[string]map[string]uint, 0)
		gasUsedLevels  = make(map[string]map[string]uint, 0)
		gasLevels      = make(map[string]map[string]uint, 0)

		transactedValues = make([]string, 0)

		contractCalls     = make([]uint64, 0)
		contractsDeployed = make([]uint64, 0)
	)

	dates := result.getDates()

	for _, date := range dates {
		e := result.getElement(date)
		elem := e.(*transactionChartData)

		transactions = append(transactions, elem.transactions)
		failedTransactions = append(failedTransactions, elem.failedTransactions)

		//TODO: check if overflows uint64
		gasUsed = append(gasUsed, elem.gasUsed.String())
		txFees = append(txFees, elem.txFees.String())

		transactedValues = append(transactedValues, elem.transactedValue.String())

		for gp, txns := range elem.gasPriceLevels {
			//Use gwei as keys
			gP, _ := new(big.Int).SetString(gp, 10)
			gP.Div(gP, big.NewInt(1000000000))
			gasPrice := gP.String()

			if _, ok := gasPriceLevels[gasPrice]; !ok {
				// When the loop encounters a new gasprice 'level', we add a new map to the map, with gadsprice as key.
				// this new map will have date as keys and txns as values
				gasPriceLevels[gasPrice] = make(map[string]uint, 1)
				gasPriceLevels[gasPrice][date] = txns
			} else {
				gasPriceLevels[gasPrice][date] = txns
			}
		}

		for gu, txns := range elem.gasUsedLevels {
			//Use gas amount as key
			gU, _ := new(big.Int).SetString(gu, 10)
			gasUsed := gU.String()

			if _, ok := gasUsedLevels[gasUsed]; !ok {
				gasUsedLevels[gasUsed] = make(map[string]uint, 1)
				gasUsedLevels[gasUsed][date] = txns
			} else {
				gasUsedLevels[gasUsed][date] = txns
			}
		}

		for gl, txns := range elem.gasLevels {
			//Use gas amount as key
			gL, _ := new(big.Int).SetString(gl, 10)
			gasLevel := gL.String()

			if _, ok := gasLevels[gasLevel]; !ok {
				gasLevels[gasLevel] = make(map[string]uint, 1)
				gasLevels[gasLevel][date] = txns
			} else {
				gasLevels[gasLevel][date] = txns
			}
		}

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

	err = c.backend.AddNumberStringChart("transactedValues", txFees, dates)
	if err != nil {
		c.logger.Error("error adding transactedValues chart", "err", err)
	}
	c.logger.Info("added transactedValues chart")

	err = c.backend.AddMultiSeriesChart("gasPriceLevels", gasPriceLevels, dates)
	if err != nil {
		c.logger.Error("error adding gasPriceLevels chart", "err", err)
	}
	c.logger.Info("added gasPriceLevels chart")

	err = c.backend.AddMultiSeriesChart("gasUsedLevels", gasUsedLevels, dates)
	if err != nil {
		c.logger.Error("error adding gasLevels chart", "err", err)
	}
	c.logger.Info("added gasLevels chart")

	err = c.backend.AddMultiSeriesChart("gasLevels", gasLevels, dates)
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
