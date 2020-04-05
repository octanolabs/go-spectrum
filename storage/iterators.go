package storage

import (
	"context"

	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) CrawlTransactions() (*mongo.Cursor, error) {

	query := []bson.M{{"$sort": bson.M{"blockNumber": 1}}}

	return m.C(models.TXNS).Find(context.Background(), query, options.Find())

}

func (m *MongoDB) CrawlBlocks() (*mongo.Cursor, error) {

	query := []bson.M{{"$sort": bson.M{"number": 1}}}

	return m.C(models.BLOCKS).Find(context.Background(), query, options.Find())

}

func (m *MongoDB) CrawlReorgs() (*mongo.Cursor, error) {

	query := []bson.M{{"$sort": bson.M{"number": 1}}}

	return m.C(models.REORGS).Find(context.Background(), query, options.Find())

}

func (m *MongoDB) CrawlUncles() (*mongo.Cursor, error) {

	query := []bson.M{{"$sort": bson.M{"blockNumber": 1}}}

	return m.C(models.UNCLES).Find(context.Background(), query, options.Find())

}

func (m *MongoDB) CrawlTokenTransfers() (*mongo.Cursor, error) {

	query := []bson.M{{"$sort": bson.M{"blockNumber": 1}}}

	return m.C(models.TRANSFERS).Find(context.Background(), query, options.Find())

}
