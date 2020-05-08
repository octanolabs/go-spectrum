package storage

import (
	"context"

	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) IterTransactions() (*mongo.Cursor, error) {

	return m.C(models.TXNS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", 1}}))

}

func (m *MongoDB) IterBlocks() (*mongo.Cursor, error) {

	return m.C(models.BLOCKS).Find(context.Background(), bson.M{}, options.Find().SetHint(bson.M{"number": 1}).SetSort(bson.D{{"number", 1}}))

}

func (m *MongoDB) IterForkedBlocks() (*mongo.Cursor, error) {

	return m.C(models.FORKEDBLOCKS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", 1}}))

}

func (m *MongoDB) IterUncles() (*mongo.Cursor, error) {

	return m.C(models.UNCLES).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", 1}}))

}

func (m *MongoDB) IterTokenTransfers() (*mongo.Cursor, error) {

	return m.C(models.TRANSFERS).Find(context.Background(), bson.M{}, options.Find().SetSort(bson.D{{"blockNumber", 1}}))

}
