package storage

import (
	"github.com/globalsign/mgo/bson"
	"github.com/octanolabs/go-spectrum/models"
)

// Store

func (m *MongoDB) Store() (models.Store, error) {
	var store models.Store

	err := m.db.C(models.STORE).Find(bson.M{}).Limit(1).One(&store)
	return store, err
}

// sblocks

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
