package storage

import (
	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"github.com/octanolabs/go-spectrum/models"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	Address  string `json:"address"`
}

type MongoDB struct {
	session *mgo.Session
	db      *mgo.Database
}

func NewConnection(cfg *Config) (*MongoDB, error) {
	session, err := mgo.DialWithInfo(&mgo.DialInfo{
		Addrs:    []string{cfg.Address},
		Database: cfg.Database,
		Username: cfg.User,
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return &MongoDB{session, session.DB("")}, nil
}

func (m *MongoDB) IsFirstRun() bool {
	var store models.Store

	err := m.db.C(models.STORE).Find(&bson.M{}).Limit(1).One(&store)

	if err != nil {
		if err.Error() == "not found" {
			return true
		} else {
			log.Fatalf("Error during initialization: %v", err)
		}
	}

	return false
}

func (m *MongoDB) IsPresent(height uint64) bool {

	if height == 0 {
		return true
	}

	if dbHead, _ := m.LatestSupplyBlock(); dbHead.Number == height {
		return false
	}

	var rbn models.RawBlockDetails
	err := m.db.C(models.BLOCKS).Find(&bson.M{"number": height}).Limit(1).One(&rbn)

	if err != nil {
		if err.Error() == "not found" {
			return false
		} else {
			log.Errorf("Error checking for block in db: %v", err)
		}
	}

	return true
}

func (m *MongoDB) IsInDB(height uint64, hash string) (bool, bool) {
	var rbn models.RawBlockDetails
	err := m.db.C(models.BLOCKS).Find(&bson.M{"number": height}).Limit(1).One(&rbn)

	if err != nil {
		if err.Error() == "not found" {
			return false, false
		} else {
			log.Errorf("Error checking for block in db: %v", err)
		}
	}

	if _, contendentHash := rbn.Convert(); contendentHash != hash {
		return true, true
	}

	return true, false
}

func (m *MongoDB) Purge(height uint64) {

	// TODO: make this better

	blockselector := &bson.M{"number": height}

	bulk := m.db.C(models.SBLOCK).Bulk()
	bulk.RemoveAll(blockselector)
	_, err := bulk.Run()
	if err != nil {
		log.Errorf("Error purging blocks: %v", err)
	}

}

func (m *MongoDB) Ping() error {
	return m.session.Ping()
}
