package storage

import (
	"context"

	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//TODO:
//	Add new getters
// -failed transactions (status: false)
// -contract calls

// Store

func (m *MongoDB) Status() (models.Store, error) {
	var store models.Store

	sr := m.C(models.STORE).FindOne(context.Background(), bson.M{"symbol": m.symbol}, options.FindOne())

	err := sr.Decode(&store)

	return store, err
}

// Blocks

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

func (m *MongoDB) TotalUncleCount() (int64, error) {
	count, err := m.C(models.UNCLES).CountDocuments(context.Background(), bson.M{}, options.Count())

	return count, err
}

// Forked blocks

func (m *MongoDB) ForkedBlockByNumber(number uint64) (models.Block, error) {
	var block models.Block

	err := m.C(models.FORKEDBLOCKS).FindOne(context.Background(), bson.M{"number": number}, options.FindOne()).Decode(&block)
	return block, err
}

func (m *MongoDB) TotalForkedBlockCount() (int64, error) {
	count, err := m.C(models.FORKEDBLOCKS).CountDocuments(context.Background(), bson.M{}, options.Count())

	return count, err
}

// Transactions
func (m *MongoDB) TransactionByHash(hash string) (models.Transaction, error) {
	var txn models.Transaction

	err := m.C(models.TXNS).FindOne(context.Background(), bson.M{"hash": hash}, options.FindOne()).Decode(&txn)
	return txn, err
}

func (m *MongoDB) TransactionsByBlockNumber(number uint64) ([]models.Transaction, error) {
	var txns = make([]models.Transaction, 0)

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{"blockNumber": number}, options.Find())

	if err != nil {
		return txns, err
	}

	err = c.All(context.Background(), &txns)

	return txns, err
}

func (m *MongoDB) TransactionByContractAddress(address string) (models.Transaction, error) {
	var txn models.Transaction

	err := m.C(models.TXNS).FindOne(context.Background(), bson.M{"contractAddress": address}, options.FindOne()).Decode(&txn)
	return txn, err
}

func (m *MongoDB) TxnCount(address string) (int64, error) {
	count, err := m.C(models.TXNS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"from": address}, {"to": address}}}, options.Count())
	return count, err
}

func (m *MongoDB) TotalTxnCount() (int64, error) {
	count, err := m.C(models.TXNS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}

// Tx trace

func (m *MongoDB) TxTrace(hash string) (models.ITransaction, error) {
	var itxn models.ITransaction

	c := m.C(models.INTERNALTXNS).FindOne(context.Background(), bson.M{"parentHash": hash}, options.FindOne())

	err := c.Decode(&itxn)
	if err != nil {
		return models.ITransaction{}, err
	}

	return itxn, err
}

func (m *MongoDB) LatestTxTrace() (models.ITransaction, error) {
	var trace models.ITransaction

	err := m.C(models.INTERNALTXNS).FindOne(context.Background(), bson.M{}, options.FindOne().SetSort(bson.D{{"blockNumber", -1}})).Decode(&trace)
	return trace, err
}

// Contracts

func (m *MongoDB) TotalContractCallsCount() (int64, error) {
	count, err := m.C(models.CONTRACTCALLS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}

func (m *MongoDB) TotalContractsDeployedCount() (int64, error) {
	count, err := m.C(models.CONTRACTS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}

// Token transfers

func (m *MongoDB) TransfersOfTokenByAccount(token string, account string) ([]models.TokenTransfer, error) {
	var transfers = make([]models.TokenTransfer, 0)

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"$or": []bson.M{{"$and": []bson.M{{"from": account}, {"contract": token}}}, {"$and": []bson.M{{"to": account}, {"contract": token}}}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) TransfersOfTokenByAccountCount(token string, account string) (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"$and": []bson.M{{"from": account}, {"contract": token}}}, {"$and": []bson.M{{"to": account}, {"contract": token}}}}}, options.Count())
	return count, err
}

func (m *MongoDB) TokenTransfersByAccount(account string) ([]models.TokenTransfer, error) {
	var transfers = make([]models.TokenTransfer, 0)

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"$or": []bson.M{{"from": account}, {"to": account}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) TokenTransfersByAccountCount(account string) (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"$or": []bson.M{{"from": account}, {"to": account}}}, options.Count())
	return count, err
}

func (m *MongoDB) TransfersByContract(address string) ([]models.TokenTransfer, error) {
	var transfers = make([]models.TokenTransfer, 0)

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"contract": address}, options.Find().SetSort(bson.D{{"blockNumber", -1}}))

	if err != nil {
		return transfers, err
	}

	err = c.All(context.Background(), &transfers)

	return transfers, err
}

func (m *MongoDB) ContractTransferCount(address string) (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{"contract": address}, options.Count())
	return count, err
}

func (m *MongoDB) TotalTransferCount() (int64, error) {
	count, err := m.C(models.TRANSFERS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}

// Charts
// TODO: use multikey indexes to return part of the data

func (m *MongoDB) GetNumberChart(name string, limit int) (models.NumberChart, error) {
	var chart models.NumberChart

	err := m.C(models.CHARTS).FindOne(context.Background(), bson.M{"name": name}, options.FindOne()).Decode(&chart)
	if err != nil {
		return models.NumberChart{}, err
	}

	if limit > 0 {
		lastIdx := len(chart.Series) - 1
		chart.Series = chart.Series[lastIdx-limit:]
		chart.Timestamps = chart.Timestamps[lastIdx-limit:]
	}

	return chart, err
}

func (m *MongoDB) GetNumberStringChart(name string, limit int) (models.NumberStringChart, error) {
	var chart models.NumberStringChart

	err := m.C(models.CHARTS).FindOne(context.Background(), bson.M{"name": name}, options.FindOne()).Decode(&chart)
	if err != nil {
		return models.NumberStringChart{}, err
	}

	if limit > 0 {
		lastIdx := len(chart.Series) - 1
		chart.Series = chart.Series[lastIdx-limit:]
		chart.Timestamps = chart.Timestamps[lastIdx-limit:]
	}

	return chart, err
}

func (m *MongoDB) GetMultiSeriesChart(name string, limit int) (models.MultiSeriesChart, error) {
	var chart models.MultiSeriesChart

	err := m.C(models.CHARTS).FindOne(context.Background(), bson.M{"name": name}, options.FindOne()).Decode(&chart)
	if err != nil {
		return models.MultiSeriesChart{}, err
	}

	if limit > 0 {
		for _, series := range chart.Datasets {
			series.SliceTime(limit)
		}
		chart.Timestamps = chart.Timestamps[len(chart.Timestamps)-limit : len(chart.Timestamps)-1]
	}

	return chart, err
}

func (m *MongoDB) ListCharts() ([]string, error) {
	var charts []struct {
		Name string `json:"name" bson:"name"`
	}
	var result []string

	c, err := m.C(models.CHARTS).Find(context.Background(), bson.M{}, options.Find().SetProjection(bson.M{"name": 1}))
	if err != nil {
		return nil, err
	}

	err = c.All(context.Background(), &charts)
	if err != nil {
		return nil, err
	}

	for _, v := range charts {
		result = append(result, v.Name)
	}

	return result, err
}

// Accounts

func (m *MongoDB) TotalAccountCount() (int64, error) {
	count, err := m.C(models.ACCOUNTS).CountDocuments(context.Background(), bson.M{}, options.Count())
	return count, err
}
