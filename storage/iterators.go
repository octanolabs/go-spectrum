package storage

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) IterTransactions(from, to int64) (*mongo.Cursor, error) {

	query := bson.M{"$and": []bson.M{{"timestamp": bson.M{"$gt": from}}, {"timestamp": bson.M{"$lt": to}}}}

	return m.C(models.TXNS).Find(context.Background(), query, options.Find().SetHint(bson.M{"blockNumber": 1}).SetSort(bson.D{{"blockNumber", 1}}))
}

func (m *MongoDB) IterBlocks(from, to int64) (*mongo.Cursor, error) {

	query := bson.M{"$and": []bson.M{{"timestamp": bson.M{"$gt": from}}, {"timestamp": bson.M{"$lt": to}}}}

	return m.C(models.BLOCKS).Find(context.Background(), query, options.Find().SetHint(bson.M{"number": 1}).SetSort(bson.D{{"number", 1}}))

}

func (m *MongoDB) IterForkedBlocks(from, to int64) (*mongo.Cursor, error) {

	query := bson.M{"$and": []bson.M{{"timestamp": bson.M{"$gt": from}}, {"timestamp": bson.M{"$lt": to}}}}

	return m.C(models.FORKEDBLOCKS).Find(context.Background(), query, options.Find().SetHint(bson.M{"number": 1}).SetSort(bson.D{{"number", 1}}))
}

func (m *MongoDB) IterUncles(from, to int64) (*mongo.Cursor, error) {

	query := bson.M{"$and": []bson.M{{"timestamp": bson.M{"$gt": from}}, {"timestamp": bson.M{"$lt": to}}}}

	return m.C(models.UNCLES).Find(context.Background(), query, options.Find().SetHint(bson.M{"blockNumber": 1}).SetSort(bson.D{{"blockNumber", 1}}))
}

func (m *MongoDB) IterTokenTransfers(from, to int64) (*mongo.Cursor, error) {
	query := bson.M{"$and": []bson.M{{"timestamp": bson.M{"$gt": from}}, {"timestamp": bson.M{"$lt": to}}}}

	return m.C(models.TRANSFERS).Find(context.Background(), query, options.Find().SetHint(bson.M{"blockNumber": 1}).SetSort(bson.D{{"blockNumber", 1}}))
}
