package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/octanolabs/go-spectrum/models"
	"github.com/ubiq/go-ubiq/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Symbol   string `json:"symbol"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Address  string `json:"address"`
}

func (c *Config) ConnectionString() string {
	return fmt.Sprint("mongodb://", c.User, ":", c.Password, "@", c.Address, "/", c.Database)
}

type MongoDB struct {
	symbol string
	client *mongo.Client
	db     *mongo.Database
}

func NewConnection(cfg *Config) (*MongoDB, error) {

	if cfg.Symbol == "" {
		return nil, errors.New("symbol not set")
	}

	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.ConnectionString()))
	if err != nil {
		log.Error("error creating mongo client", "err", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Error("couldn't connect to mongo", "err", err)
	}

	return &MongoDB{cfg.Symbol, client, client.Database(cfg.Database, options.Database())}, nil
}

func (m *MongoDB) C(coll string) *mongo.Collection {
	return m.db.Collection(coll, options.Collection())
}

func (m *MongoDB) IsFirstRun() bool {

	err := m.C(models.STORE).FindOne(context.Background(), bson.M{}, options.FindOne()).Err()

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true
		} else {
			log.Error("Error during initialization", "err", err)
		}
	}

	return false
}

func (m *MongoDB) Ping() error {
	return m.client.Ping(context.Background(), nil)
}

func (m *MongoDB) PurgeBlock(height uint64) error {

	r, err := m.C(models.BLOCKS).DeleteOne(context.Background(), bson.M{"number": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debug("purged %v blocks", "count", r.DeletedCount)

	r, err = m.C(models.TXNS).DeleteMany(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debug("purged %v transactions", "count", r.DeletedCount)

	r, err = m.C(models.TRANSFERS).DeleteMany(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debug("purged %v transfers", "count", r.DeletedCount)
	return nil

}

func (m *MongoDB) IsEnodePresent(id string) bool {

	err := m.C(models.ENODES).FindOne(context.Background(), bson.M{"id": id}, options.FindOne()).Err()

	if err != nil {
		log.Debug("could not find enode", "err", err)
		return false
	}
	return true
}

func (m *MongoDB) UpdateStore() error {

	var (
		txCount, transferCount, uncleCount int64
		latestBlock                        models.Block
	)

	collection := m.C(models.STORE)

	txCount, err := m.TotalTxnCount()

	if err != nil {
		return err
	}

	transferCount, err = m.TotalTransferCount()

	if err != nil {
		return err
	}

	uncleCount, err = m.TotalUncleCount()

	if err != nil {
		return err
	}

	latestBlock, err = m.LatestBlock()

	if err != nil {
		return err
	}

	filter := bson.M{"symbol": m.symbol}
	update := bson.D{{"$set", bson.M{
		"updated":             time.Now().Unix(),
		"supply":              latestBlock.Supply,
		"latestBlock":         latestBlock,
		"totalTransactions":   txCount,
		"totalTokenTransfers": transferCount,
		"totalUncles":         uncleCount,
	}}}

	updateRes, err := collection.UpdateOne(context.Background(), filter, update, options.Update())
	if err != nil {
		return err
	}

	if updateRes.ModifiedCount == 0 {
		return errors.New("didn't update " + m.symbol + " store")
	}
	return nil

}
