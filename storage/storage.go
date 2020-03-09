package storage

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/octanolabs/go-spectrum/models"
	log "github.com/sirupsen/logrus"
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
		log.Fatalf("Error creating mongo client: %v", err)
		log.Debugf("Unwrapped: %v", errors.Unwrap(err))
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)

	defer cancel()

	err = client.Connect(ctx)

	if err != nil {
		log.Fatalf("Error connecting to mongo: %v", err)
		log.Debugf("Unwrapped: %v", errors.Unwrap(err))
	}

	return &MongoDB{cfg.Symbol, client, client.Database(cfg.Database, options.Database())}, nil
}

func (m *MongoDB) C(coll string) *mongo.Collection {
	return m.db.Collection(coll, options.Collection())
}

func (m *MongoDB) IsFirstRun() bool {

	err := m.C(models.STORE).FindOne(context.Background(), bson.M{}, options.FindOne()).Err()

	log.Debugf("err: %#v", err)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return true
		} else {
			log.Fatalf("Error during initialization: %v", err)
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
	log.Debugf("Purged %v blocks", r.DeletedCount)

	r, err = m.C(models.TXNS).DeleteMany(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debugf("Purged %v transactions", r.DeletedCount)

	r, err = m.C(models.TRANSFERS).DeleteMany(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debugf("Purged %v transfers", r.DeletedCount)
	return nil

}

func (m *MongoDB) IsEnodePresent(id string) bool {

	err := m.C(models.ENODES).FindOne(context.Background(), bson.M{"id": id}, options.FindOne()).Err()

	if err != nil {
		log.Debugf("Error: could not find enode: %#v", err)
		return false
	}
	return true
}
