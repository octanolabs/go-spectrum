package storage

import (
	"github.com/globalsign/mgo/bson"
	"github.com/octanolabs/go-spectrum/models"
)

// Getters

func (m *MongoDB) SupplyObject(symbol string) (models.Store, error) {
	var store models.Store

	err := m.db.C(models.STORE).Find(bson.M{"symbol": symbol}).One(&store)
	return store, err
}

func (m *MongoDB) LatestSupplyBlock() (models.Sblock, error) {
	var block models.Sblock

	err := m.db.C(models.SBLOCK).Find(bson.M{}).Sort("-number").Limit(1).One(&block)
	return block, err
}

func (m *MongoDB) SupplyBlockByNumber(number uint64) (*models.Sblock, error) {
	var block *models.Sblock

	err := m.db.C(models.SBLOCK).Find(bson.M{"number": number}).One(&block)
	return block, err
}

func (m *MongoDB) SupplyBlockByHash(hash string) (*models.Sblock, error) {
	var block *models.Sblock

	err := m.db.C(models.SBLOCK).Find(bson.M{"hash": hash}).One(&block)
	return block, err
}

// Setters

func (m *MongoDB) RemoveSupplyBlock(height uint64) error {
	selector := &bson.M{"number": height}

	bulk := m.db.C(models.SBLOCK).Bulk()
	bulk.RemoveAll(selector)
	_, err := bulk.Run()

	return err
}

func (m *MongoDB) AddSupplyBlock(b models.Sblock) error {
	ss := m.db.C(models.SBLOCK)

	if err := ss.Insert(b); err != nil {
		return err
	}
	return nil
}
