package storage

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"

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

func (m *MongoDB) AddDeployedContract(tx *models.Transaction) error {
	collection := m.C(models.CONTRACTS)

	if _, err := collection.InsertOne(context.Background(), tx, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddContractCall(tx *models.Transaction) error {
	collection := m.C(models.CONTRACTCALLS)

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
	collection := m.C(models.FORKEDBLOCKS)

	if _, err := collection.InsertOne(context.Background(), b, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddEnodes(e *models.Enode) error {
	collection := m.C(models.ENODES)

	if _, err := collection.InsertOne(context.Background(), e, options.InsertOne()); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddNumberChart(name string, series []uint64, stamps []string) error {
	collection := m.C(models.CHARTS)

	if _, err := collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.D{{"$set", &models.NumberChart{
		Name:       name,
		Series:     series,
		Timestamps: stamps,
	}}}, options.Update().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddNumberStringChart(name string, series []string, stamps []string) error {
	collection := m.C(models.CHARTS)

	if _, err := collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.D{{"$set", &models.NumberStringChart{
		Name:       name,
		Series:     series,
		Timestamps: stamps,
	}}}, options.Update().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}

func (m *MongoDB) AddMLChart(name string, series interface{}, stamps []string) error {
	collection := m.C(models.CHARTS)

	if _, err := collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.D{{"$set", &models.MLChart{
		Name:       name,
		Series:     series,
		Timestamps: stamps,
	}}}, options.Update().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}
