package storage

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// These methods are used exclusively by the api, and since they return a subset of elements in a given collection we also include the totals for those collections

func (m *MongoDB) LatestBlocks(limit int64) (map[string]interface{}, error) {
	var (
		blocks = make([]models.Block, 0)
		result = map[string]interface{}{}
	)

	c, err := m.C(models.BLOCKS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(limit))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &blocks)

	result["blocks"] = blocks
	result["total"] = blocks[0].Number + 1

	return result, err
}

func (m *MongoDB) LatestUncles(limit int64) (map[string]interface{}, error) {
	var (
		uncles = make([]models.Uncle, 0)
		result = map[string]interface{}{}
	)

	status, err := m.Status()

	if err != nil {
		return result, err
	}

	c, err := m.C(models.UNCLES).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &uncles)

	result["uncles"] = uncles
	result["total"] = status.TotalUncles

	return result, err
}

func (m *MongoDB) LatestForkedBlocks(limit int64) (map[string]interface{}, error) {
	var (
		forkedBlocks = make([]models.Block, 0)
		result       = map[string]interface{}{}
	)

	status, err := m.Status()
	if err != nil {
		return result, err
	}

	c, err := m.C(models.FORKEDBLOCKS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"number", -1}}).SetLimit(limit))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &forkedBlocks)

	result["forkedBlocks"] = forkedBlocks
	result["total"] = status.TotalForkedBlocks

	return result, err
}

func (m *MongoDB) LatestTransactions(limit int64) (map[string]interface{}, error) {
	var (
		txns   = make([]models.Transaction, 0)
		result = map[string]interface{}{}
	)

	status, err := m.Status()
	if err != nil {
		return result, err
	}

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &txns)

	result["txns"] = txns
	result["total"] = status.TotalTransactions

	return result, err
}

func (m *MongoDB) LatestTokenTransfers(limit int64) (map[string]interface{}, error) {
	var (
		transfers = make([]models.TokenTransfer, 0)
		result    = map[string]interface{}{}
	)

	status, err := m.Status()
	if err != nil {
		return result, err
	}

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(limit))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &transfers)

	result["transfers"] = transfers
	result["total"] = status.TotalTokenTransfers

	return result, err
}

func (m *MongoDB) LatestTransfersOfToken(hash string) (map[string]interface{}, error) {
	var (
		transfers = make([]models.TokenTransfer, 0)
		result    = map[string]interface{}{}
	)

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"contract": hash}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(1000))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &transfers)

	count, err := m.ContractTransferCount(hash)

	if err != nil {
		return result, err
	}

	result["transfers"] = transfers
	result["total"] = count

	return result, err
}

func (m *MongoDB) LatestTransactionsByAccount(hash string) (map[string]interface{}, error) {
	var (
		txns   = make([]models.Transaction, 0)
		result = map[string]interface{}{}
	)

	c, err := m.C(models.TXNS).Find(context.Background(), bson.M{"$or": []bson.M{{"from": hash}, {"to": hash}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(100))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &txns)

	count, err := m.TxnCount(hash)

	if err != nil {
		return result, err
	}

	result["txns"] = txns
	result["total"] = count

	return result, err
}

func (m *MongoDB) LatestTokenTransfersByAccount(account string) (map[string]interface{}, error) {
	var (
		transfers = make([]models.TokenTransfer, 0)
		result    = map[string]interface{}{}
	)

	c, err := m.C(models.TRANSFERS).Find(context.Background(), bson.M{"$or": []bson.M{{"from": account}, {"to": account}}}, options.Find().SetSort(bson.D{{"blockNumber", -1}}).SetLimit(100))

	if err != nil {
		return result, err
	}

	err = c.All(context.Background(), &transfers)

	count, err := m.TokenTransfersByAccountCount(account)

	if err != nil {
		return result, err
	}

	result["transfers"] = transfers
	result["total"] = count

	return result, err
}
