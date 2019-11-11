package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/octanolabs/go-spectrum/models"
	log "github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Address  string `json:"address"`
}

func (c *Config) connectionString() string {
	return fmt.Sprint("mongodb://", c.User, ":", c.Password, "@", c.Address)
}

type MongoDB struct {
	client *mongo.Client
	db     *mongo.Database
}

// todo:
// √ Store
// Add back indexes
// √ continue moving methods to mongo driver
// √ switch the supply code to explorer code

func NewConnection(cfg *Config) (*MongoDB, error) {

	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.connectionString()))
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

	return &MongoDB{client, client.Database(cfg.Database, options.Database())}, nil
}

func (m *MongoDB) C(coll string) *mongo.Collection {
	return m.db.Collection(coll, options.Collection())
}

func (m *MongoDB) IsFirstRun() bool {

	err := m.C(models.STORE).FindOne(context.Background(), bson.D{}, options.FindOne()).Err()

	if err != nil {
		if err.Error() == "not found" {
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

	r, err = m.C(models.TXNS).DeleteOne(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debugf("Purged %v transactions", r.DeletedCount)

	r, err = m.C(models.TRANSFERS).DeleteOne(context.Background(), bson.M{"blockNumber": height}, options.Delete())

	if err != nil {
		return err
	}
	log.Debugf("Purged %v transfers", r.DeletedCount)
	return nil

}
