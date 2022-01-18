package storage

import (
	"context"
	"math/big"
	"sort"

	"github.com/octanolabs/go-spectrum/util"
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

func (m *MongoDB) AddTxTrace(itxn *models.TxTrace) error {
	collection := m.C(models.INTERNALTXNS)

	if _, err := collection.InsertOne(context.Background(), itxn, options.InsertOne()); err != nil {
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

func (m *MongoDB) AddAccount(a *models.Account) error {
	collection := m.C(models.ACCOUNTS)

	if _, err := collection.UpdateOne(context.Background(), bson.M{"address": a.Address}, bson.D{{"$set", &models.Account{
		Address: a.Address,
		Balance: a.Balance,
		Block:   a.Block,
	}}}, options.Update().SetUpsert(true)); err != nil {
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

func (m *MongoDB) AddMultiSeriesChart(name string, series map[string]map[string]uint, stamps []string) error {
	collection := m.C(models.CHARTS)

	datasets := make([]*models.MultiSeriesDataset, 0)

	for k, v := range series {

		series := make([]uint, 0)
		timestamps := make([]string, 0)

		dst := models.MultiSeriesDataset{
			Name: k,
		}

		for date, val := range v {
			series = append(series, val)
			timestamps = append(timestamps, date)
		}

		toSort := util.DateValuesSlice{
			Values: series,
			Dates:  timestamps,
		}

		sort.Sort(toSort)

		dst.Series = toSort.Values
		dst.Timestamps = toSort.Dates

		datasets = append(datasets, &dst)
	}

	sort.Slice(datasets, func(i, j int) bool {
		sI, ok := new(big.Int).SetString(datasets[i].Name, 10)
		if !ok {
			return false
		}
		sj, ok := new(big.Int).SetString(datasets[j].Name, 10)
		if !ok {
			return false
		}
		return sI.Cmp(sj) == -1
	})

	if _, err := collection.UpdateOne(context.Background(), bson.M{"name": name}, bson.D{{"$set", &models.MultiSeriesChart{
		Name:       name,
		Datasets:   datasets,
		Timestamps: stamps,
	}}}, options.Update().SetUpsert(true)); err != nil {
		return err
	}
	return nil
}
