package storage

import (
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/globalsign/mgo"
	"github.com/octanolabs/go-spectrum/models"
)

func (m *MongoDB) Init() {
	store := &models.Store{
		Timestamp: time.Now().Unix(),
		Symbol:    "ubq",
		Sync:      [1]uint64{1 << 62},
	}

	ss := m.db.C(models.STORE)

	if err := ss.Insert(store); err != nil {
		log.Fatalf("Could not init sysStore(ubq): %v", err)
	}

	genesis := &models.Sblock{
		Number:       0,
		Hash:         "0x406f1b7dd39fca54d8c702141851ed8b755463ab5b560e6f19b963b4047418af",
		Timestamp:    1485633600,
		BlockReward:  "0",
		UncleRewards: "0",
		Minted:       "36108073197716300000000000",
		Supply:       "36108073197716300000000000",
	}

	sb := m.db.C(models.SBLOCK)
	m.db.C(models.SBLOCK).EnsureIndex(mgo.Index{Key: []string{"-number"}, Unique: true, Background: true})

	if err := sb.Insert(genesis); err != nil {
		log.Fatalf("Could not init supply block: %v", err)
	}

	log.Warnf("Initialized sysStore, genesis")
}
