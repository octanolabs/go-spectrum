package storage

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Store

func (m *MongoDB) ChainStore(symbol string) (models.Store, error) {
	var store models.Store

	sr := m.C(models.STORE).FindOne(context.Background(), bson.M{"symbol": symbol}, options.FindOne())

	err := sr.Decode(&store)

	return store, err
}

// Getters

func (m *MongoDB) BlockByNumber(number uint64) (models.Block, error) {
	var block models.Block

	err := m.C(models.BLOCKS).FindOne(context.Background(), bson.M{"number": number}, options.FindOne()).Decode(&block)

	return block, err
}

func (m *MongoDB) BlockByHash(hash string) (models.Block, error) {
	var block models.Block

	err := m.C(models.BLOCKS).FindOne(context.Background(), bson.M{"hash": hash}, options.FindOne()).Decode(&block)
	return block, err
}

func (m *MongoDB) LatestBlock() (models.Block, error) {
	var block models.Block

	err := m.C(models.BLOCKS).FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.D{{"number", -1}})).Decode(&block)
	return block, err
}

func (m *MongoDB) LatestBlocks(limit int) ([]models.Block, error) {
	var blocks []models.Block

	c, err := m.C(models.BLOCKS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(50))

	if err != nil {
		return blocks, err
	}

	err = c.All(context.Background(), &blocks)

	return blocks, err
}

func (m *MongoDB) TotalBlockCount() (int64, error) {
	count, err := m.C(models.BLOCKS).CountDocuments(context.Background(), bson.M{}, options.Count())

	return count, err
}

// Uncles

func (m *MongoDB) UncleByHash(hash string) (models.Uncle, error) {
	var uncle models.Uncle

	err := m.C(models.UNCLES).FindOne(context.Background(), bson.M{"hash": hash}, options.FindOne()).Decode(&uncle)
	return uncle, err
}

func (m *MongoDB) LatestUncles(limit int64) ([]models.Uncle, error) {
	var uncles []models.Uncle

	c, err := m.C(models.UNCLES).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return uncles, err
	}

	err = c.All(context.Background(), &uncles)

	return uncles, err
}

func (m *MongoDB) TotalUncleCount() (int64, error) {
	count, err := m.C(models.UNCLES).CountDocuments(context.Background(), bson.M{}, options.Count())

	return count, err
}

// Forked blocks

func (m *MongoDB) ForkedBlockByNumber(number uint64) (models.Block, error) {
	var block models.Block

	err := m.C(models.REORGS).FindOne(context.Background(), bson.M{"number": number}, options.FindOne()).Decode(&block)
	return block, err
}

func (m *MongoDB) LatestForkedBlocks(limit int64) ([]models.Block, error) {
	var blocks []models.Block

	c, err := m.C(models.REORGS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(limit))

	if err != nil {
		return blocks, err
	}

	err = c.All(context.Background(), &blocks)

	return blocks, err
}

// Transactions

func (m *MongoDB) TransactionsByBlockNumber(number uint64) ([]models.Transaction, error) {
	var txns []models.Transaction

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{"blockNumber": number}, options.Find())

	if err != nil {
		return txns, err
	}

	err = c.All(context.Background(), &txns)

	return txns, err
}

func (m *MongoDB) TransactionByHash(hash string) (models.Transaction, error) {
	var txn models.Transaction

	err := m.C(models.TXNS).FindOne(context.Background(), bson.M{"hash": hash}, options.FindOne()).Decode(&txn)
	return txn, err
}

func (m *MongoDB) TransactionByContractAddress(hash string) (models.Transaction, error) {
	var txn models.Transaction

	err := m.C(models.TXNS).FindOne(context.Background(), bson.M{"contractAddress": hash}, options.FindOne()).Decode(&txn)
	return txn, err
}

func (m *MongoDB) LatestTransactions(limit int64) ([]models.Transaction, error) {
	var txns []models.Transaction

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return txns, err
	}

	err = c.All(context.Background(), &txns)

	return txns, err
}

func (m *MongoDB) LatestTransactionsByAccount(hash string) ([]models.Transaction, error) {
	var txns []models.Transaction

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{"$or": []bson.M{{"from": hash}, {"to": hash}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(100))

	if err != nil {
		return txns, err
	}

	err = c.All(context.Background(), &txns)

	return txns, err
}

func (m *MongoDB) TxnCount(hash string) (int64, error) {
	count, err := m.C(models.TXNS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"from": hash}, {"to": hash}}}, options.Count())
	return count, err
}

func (m *MongoDB) TotalTxnCount() (int64, error) {
	count, err := m.C(models.TXNS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}

// Token transfers

func (m *MongoDB) TokenTransfersByAccount(token string, account string) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"$or": []bson.M{{"$and": []bson.M{{"from": account}, {"contract": token}}}, {"$and": []bson.M{{"to": account}, {"contract": token}}}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) TokenTransfersByAccountCount(token string, account string) (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"$and": []bson.M{{"from": account}, {"contract": token}}}, {"$and": []bson.M{{"to": account}, {"contract": token}}}}}, options.Count())
	return count, err
}

func (m *MongoDB) LatestTokenTransfersByAccount(hash string) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"$or": []bson.M{{"from": hash}, {"to": hash}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(100))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) LatestTransfersOfToken(hash string) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"contract": hash}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(1000))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) TokenTransferCount(hash string) (int64, error) {
	// This functions is identical to transactionCount
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"from": hash}, {"to": hash}}}, options.Count())
	return count, err
}

func (m *MongoDB) TokenTransferCountByContract(hash string) (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"contract": hash}, options.Count())
	return count, err
}

func (m *MongoDB) LatestTokenTransfers(limit int64) ([]models.TokenTransfer, error) {
	var transfers []models.TokenTransfer

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

// Charts

//func (m *MongoDB) ChartData(chart string, limit int64) (models.LineChart, error) {
//	var chartData models.LineChart
//
//	err := m.C(models.CHARTS).Find(context.Background(), bson.M{"chart": chart}).One(&chartData)
//
//	if err != nil {
//		return models.LineChart{}, err
//	}
//
//	if limit >= int64(len(chartData.Labels)) || limit >= int64(len(chartData.Values)) || limit == 0 {
//		limit = int64(len(chartData.Labels) - 1)
//	}
//
//	// Limit selects items from the end of the slice; we exclude the last element (current day)
//	// TODO: Eventually fix this in the iterators
//
//	chartData.Labels = chartData.Labels[int64(len(chartData.Labels)-1)-limit : len(chartData.Labels)-1]
//	chartData.Values = chartData.Values[int64(len(chartData.Values)-1)-limit : len(chartData.Values)-1]
//
//	return chartData, err
//}
//
//func (m *MongoDB) ChartDataML(chart string, limit int64, miner string) (models.LineChart, error) {
//	var chartData models.MLineChart
//	var result models.LineChart
//
//	err := m.C(models.CHARTS).Find(context.Background(), bson.M{"chart": chart}).One(&chartData)
//
//	if err != nil {
//		return models.LineChart{}, err
//	}
//
//	if limit >= int64(len(chartData.Labels)) || limit >= int64(len(chartData.Values)) || limit == 0 {
//		limit = int64(len(chartData.Labels) - 1)
//	}
//
//	// Limit selects items from the end of the slice; we exclude the last element (current day)
//	// TODO: Eventually fix this in the iterators
//
//	result.Chart = miner + " hashrate"
//	result.Labels = chartData.Labels[int64(len(chartData.Labels)-1)-limit : len(chartData.Labels)-1]
//
//	if _, ok := chartData.Values[miner]; !ok {
//		return models.LineChart{}, errors.New("Miner not found")
//	}
//
//	result.Values = chartData.Values[miner][int64(len(chartData.Values[miner])-1)-limit : len(chartData.Values[miner])-1]
//
//	return result, err
//}
