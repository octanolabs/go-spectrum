package storage

import (
	"context"
	"os"

	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/util"
	"github.com/ubiq/go-ubiq/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) Init() {
	genesis := &models.Block{
		Number:          0,
		Timestamp:       1485633600,
		Txs:             0,
		Hash:            "0x406f1b7dd39fca54d8c702141851ed8b755463ab5b560e6f19b963b4047418af",
		ParentHash:      "0x0000000000000000000000000000000000000000000000000000000000000000",
		Sha3Uncles:      "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
		Miner:           "0x3333333333333333333333333333333333333333",
		Difficulty:      "80000000000",
		TotalDifficulty: "80000000000",
		Size:            524,
		GasUsed:         0,
		GasLimit:        134217728,
		Nonce:           "0x0000000000000888",
		UncleNo:         0,
		// Empty
		BlockReward:  "0",
		UncleRewards: "0",
		AvgGasPrice:  "0",
		TxFees:       "0",
		//
		ExtraData: "0x4a756d6275636b734545",
		Minted:    "36108073197716300000000000",
		Supply:    "36108073197716300000000000",
	}

	collection := m.C(models.BLOCKS)

	if _, err := collection.InsertOne(context.Background(), genesis); err != nil {
		log.Error("could not init supply block", "err", err)
		os.Exit(1)
	}

	collection = m.C(models.STORE)

	store := &models.Store{
		Timestamp: util.MakeTimestamp(),
		Symbol:    m.symbol,
		Supply:    "36108073197716300000000000",
	}

	if _, err := collection.InsertOne(context.Background(), store); err != nil {
		log.Error("could not init supply block", "err", err)
	}

	m.initIndexes()

	log.Warn("initialized sysStore, genesis, indexes")
}

func (m *MongoDB) initIndexes() {

	iv := m.C(models.BLOCKS).Indexes()

	bnIdxModel := mongo.IndexModel{bson.M{"number": 1}, options.Index().SetName("blocksNumberIndex").SetUnique(true).SetBackground(true)}
	bhIdxModel := mongo.IndexModel{bson.M{"hash": 1}, options.Index().SetName("blocksHashIndex").SetUnique(true).SetBackground(true)}

	_, err := iv.CreateMany(context.Background(), []mongo.IndexModel{bnIdxModel, bhIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for blocks", "err", err)
	}

	iv = m.C(models.REORGS).Indexes()

	rIdxModel := mongo.IndexModel{bson.M{"hash": 1}, options.Index().SetName("reorgsIndex").SetUnique(true).SetBackground(true)}

	_, err = iv.CreateOne(context.Background(), rIdxModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init index", "name", rIdxModel.Options.Name, "err", err)
	}

	iv = m.C(models.UNCLES).Indexes()

	uIdxModel := mongo.IndexModel{bson.M{"hash": 1}, options.Index().SetName("unclesIndex").SetUnique(true).SetBackground(true)}

	_, err = iv.CreateOne(context.Background(), uIdxModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init index", "name", uIdxModel.Options.Name, "err", err)
	}

	iv = m.C(models.TXNS).Indexes()

	txHIdxModel := mongo.IndexModel{bson.M{"hash": 1}, options.Index().SetName("txHashIndex").SetUnique(true).SetBackground(true)}
	txBNIdxModel := mongo.IndexModel{bson.M{"blockNumber": 1}, options.Index().SetName("txBlockNumberIndex").SetBackground(true)}
	txFIdxModel := mongo.IndexModel{bson.M{"from": 1}, options.Index().SetName("txFromIndex").SetBackground(true)}
	txTIdxModel := mongo.IndexModel{bson.M{"to": 1}, options.Index().SetName("txToIndex").SetBackground(true)}
	txCAIdxModel := mongo.IndexModel{bson.M{"contractAddress": 1}, options.Index().SetName("txContractAddressIndex").SetBackground(true)}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{txBNIdxModel, txHIdxModel, txFIdxModel, txTIdxModel, txCAIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for transactions", "err", err)
	}

	iv = m.C(models.TRANSFERS).Indexes()

	trBNIdxModel := mongo.IndexModel{bson.M{"blockNumber": 1}, options.Index().SetName("trBlockNumberIndex").SetBackground(true)}
	trHIdxModel := mongo.IndexModel{bson.M{"hash": 1}, options.Index().SetName("trTxHashIndex").SetBackground(true)}
	trFIdxModel := mongo.IndexModel{bson.M{"from": 1}, options.Index().SetName("trFromIndex").SetBackground(true)}
	trTIdxModel := mongo.IndexModel{bson.M{"to": 1}, options.Index().SetName("trToIndex").SetBackground(true)}
	trCIdxModel := mongo.IndexModel{bson.M{"contract": 1}, options.Index().SetName("trContractIndex").SetBackground(true)}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{trBNIdxModel, trHIdxModel, trFIdxModel, trTIdxModel, trCIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for transfers", "err", err)
	}

	iv = m.C(models.ENODES).Indexes()

	enodeModel := mongo.IndexModel{bson.M{"raw_enode": 1}, options.Index().SetName("enodeIndex").SetUnique(true).SetBackground(true)}

	_, err = iv.CreateOne(context.Background(), enodeModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for enodes", "err", err)
	}

	log.Warn("initialised database indexes")

}
