package storage

import (
	"context"
	"github.com/octanolabs/go-spectrum/models"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) AddTransaction(tx *models.Transaction) error {
	collection := m.C(models.TXNS)

	if _, err := collection.InsertOne(context.Background(), tx, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddTokenTransfer(tt *models.TokenTransfer) error {
	collection := m.C(models.TRANSFERS)

	if _, err := collection.InsertOne(context.Background(), tt, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddUncle(u *models.Uncle) error {
	collection := m.C(models.UNCLES)

	if _, err := collection.InsertOne(context.Background(), u, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddBlock(b *models.Block) error {
	collection := m.C(models.BLOCKS)

	if _, err := collection.InsertOne(context.Background(), b, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddForkedBlock(b *models.Block) error {
	collection := m.C(models.REORGS)

	if _, err := collection.InsertOne(context.Background(), b, options.InsertOne()); err != nil {
		return err
	}
	return nil
}
