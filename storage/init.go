package storage

import (
	"context"
	"fmt"
	"math/big"
	"os"

	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/util"
	"github.com/ubiq/go-ubiq/v6/log"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func (m *MongoDB) Init(rpc *rpc.RPCClient) {

	// get genesis state, update seeded accounts & calculate initial supply
	collection := m.C(models.ACCOUNTS)

	initialSupply := new(big.Int).SetUint64(0)

	state, _ := rpc.GetState(0)
	for k, v := range state.Accounts {
		switch a := v.(type) {
		case map[string]interface{}:
			// add accounts balance to initial supply
			balanceStr := fmt.Sprintf("%v", a["balance"])
			balance, _ := new(big.Int).SetString(balanceStr, 10)
			initialSupply = initialSupply.Add(initialSupply, balance)
			// save account to db
			account := models.Account{Address: k, Balance: balanceStr}
			if _, err := collection.UpdateOne(context.Background(), bson.M{"address": account.Address}, bson.D{{"$set", &models.Account{
				Address: account.Address,
				Balance: account.Balance,
				Block:   0,
			}}}, options.Update().SetUpsert(true)); err != nil {
				log.Error("couldn't add account", "err", err, "address", k)
			}
		default:
			// do nothing
		}
	}

	genesis, err := rpc.GetBlockByHeight(0)
	if err != nil {
		log.Error("could not retrieve genesis block", "err", err)
		os.Exit(1)
	}

	genesis.BlockReward = "0"
	genesis.UncleRewards = "0"
	genesis.AvgGasPrice = "0"
	genesis.TxFees = "0"
	genesis.Minted = initialSupply.String()
	genesis.Supply = initialSupply.String() // "36108073197716300000000000"
	genesis.Burned = "0"
	genesis.TotalBurned = "0"
	genesis.Txs = 0

	collection = m.C(models.BLOCKS)

	if _, err := collection.InsertOne(context.Background(), genesis); err != nil {
		log.Error("could not init supply block", "err", err)
		os.Exit(1)
	}

	collection = m.C(models.STORE)

	store := &models.Store{
		Timestamp:           util.MakeTimestamp(),
		Symbol:              m.symbol,
		Supply:              initialSupply.String(),
		TotalTransactions:   0,
		TotalTokenTransfers: 0,
		TotalUncles:         0,
	}

	if _, err := collection.InsertOne(context.Background(), store); err != nil {
		log.Error("could not init supply block", "err", err)
	}

	m.initIndexes()
}

func (m *MongoDB) initIndexes() {

	iv := m.C(models.BLOCKS).Indexes()

	bnIdxModel := mongo.IndexModel{Keys: bson.M{"number": 1}, Options: options.Index().SetName("blocksNumberIndex").SetUnique(true)}
	bhIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("blocksHashIndex").SetUnique(true)}
	minerIdxModel := mongo.IndexModel{Keys: bson.M{"miner": 1}, Options: options.Index().SetName("blocksMinerIndex")}

	_, err := iv.CreateMany(context.Background(), []mongo.IndexModel{bnIdxModel, bhIdxModel, minerIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for blocks", "err", err)
	}

	iv = m.C(models.ACCOUNTS).Indexes()
	accountAddressIdxModel := mongo.IndexModel{Keys: bson.M{"address": 1}, Options: options.Index().SetName("accountsAddressIndex").SetUnique(true)}
	_, err = iv.CreateOne(context.Background(), accountAddressIdxModel, options.CreateIndexes())

	iv = m.C(models.FORKEDBLOCKS).Indexes()

	rIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("reorgsIndex").SetUnique(true)}

	_, err = iv.CreateOne(context.Background(), rIdxModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init index", "name", rIdxModel.Options.Name, "err", err)
	}

	iv = m.C(models.UNCLES).Indexes()

	uIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("unclesIndex").SetUnique(true)}

	_, err = iv.CreateOne(context.Background(), uIdxModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init index", "name", uIdxModel.Options.Name, "err", err)
	}

	iv = m.C(models.TXNS).Indexes()

	txHIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("txHashIndex").SetUnique(true)}
	txBNIdxModel := mongo.IndexModel{Keys: bson.M{"blockNumber": 1}, Options: options.Index().SetName("txBlockNumberIndex")}
	txFIdxModel := mongo.IndexModel{Keys: bson.M{"from": 1}, Options: options.Index().SetName("txFromIndex")}
	txTIdxModel := mongo.IndexModel{Keys: bson.M{"to": 1}, Options: options.Index().SetName("txToIndex")}
	txCAIdxModel := mongo.IndexModel{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetName("txContractAddressIndex")}
	txFailedIdxModel := mongo.IndexModel{Keys: bson.M{"status": 1}, Options: options.Index().SetName("txFailedIndex")}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{txBNIdxModel, txHIdxModel, txFIdxModel, txTIdxModel, txCAIdxModel, txFailedIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for transactions", "err", err)
	}

	iv = m.C(models.INTERNALTXNS).Indexes()

	internalTxnModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("txHashIndex").SetUnique(true)}

	_, err = iv.CreateOne(context.Background(), internalTxnModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for internal tx", "err", err)
	}

	iv = m.C(models.CONTRACTS).Indexes()

	contractsHIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("contractHashIndex").SetUnique(true)}
	contractsBNIdxModel := mongo.IndexModel{Keys: bson.M{"blockNumber": 1}, Options: options.Index().SetName("contractBlockNumberIndex")}
	contractsFIdxModel := mongo.IndexModel{Keys: bson.M{"from": 1}, Options: options.Index().SetName("contractFromIndex")}
	contractsTIdxModel := mongo.IndexModel{Keys: bson.M{"to": 1}, Options: options.Index().SetName("contractToIndex")}
	contractsCAIdxModel := mongo.IndexModel{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetName("contractContractAddressIndex")}
	contractsFailedIdxModel := mongo.IndexModel{Keys: bson.M{"status": 1}, Options: options.Index().SetName("contractFailedIndex")}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{contractsHIdxModel, contractsBNIdxModel, contractsFIdxModel, contractsTIdxModel, contractsCAIdxModel, contractsFailedIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for contracts", "err", err)
	}

	iv = m.C(models.CONTRACTCALLS).Indexes()

	contractCallsHIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("contractCallsHashIndex").SetUnique(true)}
	contractCallsBNIdxModel := mongo.IndexModel{Keys: bson.M{"blockNumber": 1}, Options: options.Index().SetName("contractCallsBlockNumberIndex")}
	contractCallsFIdxModel := mongo.IndexModel{Keys: bson.M{"from": 1}, Options: options.Index().SetName("contractCallsFromIndex")}
	contractCallsTIdxModel := mongo.IndexModel{Keys: bson.M{"to": 1}, Options: options.Index().SetName("contractCallsToIndex")}
	contractCallsCAIdxModel := mongo.IndexModel{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetName("contractCallsContractAddressIndex")}
	contractCallsFailedIdxModel := mongo.IndexModel{Keys: bson.M{"status": 1}, Options: options.Index().SetName("contractCallsFailedIndex")}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{contractCallsHIdxModel, contractCallsBNIdxModel, contractCallsFIdxModel, contractCallsTIdxModel, contractCallsCAIdxModel, contractCallsFailedIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for contract calls", "err", err)
	}

	iv = m.C(models.TRANSFERS).Indexes()

	trBNIdxModel := mongo.IndexModel{Keys: bson.M{"blockNumber": 1}, Options: options.Index().SetName("trBlockNumberIndex")}
	trHIdxModel := mongo.IndexModel{Keys: bson.M{"hash": 1}, Options: options.Index().SetName("trTxHashIndex")}
	trFIdxModel := mongo.IndexModel{Keys: bson.M{"from": 1}, Options: options.Index().SetName("trFromIndex")}
	trTIdxModel := mongo.IndexModel{Keys: bson.M{"to": 1}, Options: options.Index().SetName("trToIndex")}
	trCIdxModel := mongo.IndexModel{Keys: bson.M{"contractAddress": 1}, Options: options.Index().SetName("trContractIndex")}
	trFailedIdxModel := mongo.IndexModel{Keys: bson.M{"status": 1}, Options: options.Index().SetName("txFailedIndex")}

	_, err = iv.CreateMany(context.Background(), []mongo.IndexModel{trBNIdxModel, trHIdxModel, trFIdxModel, trTIdxModel, trCIdxModel, trFailedIdxModel}, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for transfers", "err", err)
	}

	iv = m.C(models.ENODES).Indexes()

	enodeModel := mongo.IndexModel{Keys: bson.M{"raw_enode": 1}, Options: options.Index().SetName("enodeIndex").SetUnique(true)}

	_, err = iv.CreateOne(context.Background(), enodeModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for enodes", "err", err)
	}

	iv = m.C(models.CHARTS).Indexes()

	chartsModel := mongo.IndexModel{Keys: bson.M{"name": 1}, Options: options.Index().SetName("nameIndex").SetUnique(true)}

	_, err = iv.CreateOne(context.Background(), chartsModel, options.CreateIndexes())

	if err != nil {
		log.Error("could not init indexes for enodes", "err", err)
	}

	log.Warn("initialised database indexes")

}
