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
