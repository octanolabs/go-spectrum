package models

import (
	"net"

	"github.com/ubiq/go-ubiq/p2p/enode"
)

const (
	BLOCKS        = "blocks"
	TXNS          = "transactions"
	UNCLES        = "uncles"
	CONTRACTS     = "contracts"
	CONTRACTCALLS = "contractcalls"
	TRANSFERS     = "tokentransfers"
	FORKEDBLOCKS  = "forkedblocks"
	CHARTS        = "charts"
	STORE         = "sysstores"
	ENODES        = "enodes"
)

type Store struct {
	Timestamp   int64  `bson:"timestamp" json:"timestamp"`
	Symbol      string `bson:"symbol" json:"symbol"`
	Supply      string `bson:"supply" json:"supply"`
	LatestBlock Block  `bson:"latestBlock" json:"latestBlock"`
	Price       string `bson:"price" json:"price"`

	TotalTransactions      int64 `bson:"totalTransactions" json:"totalTransactions"`
	TotalContractsDeployed int64 `bson:"totalContractsDeployed" json:"totalContractsDeployed"`
	TotalContractCalls     int64 `bson:"totalContractCalls" json:"totalContractCalls"`
	TotalTokenTransfers    int64 `bson:"totalTokenTransfers" json:"totalTokenTransfers"`
	TotalForkedBlocks      int64 `bson:"totalForkedBlocks" json:"totalForkedBlocks"`
	TotalUncles            int64 `bson:"totalUncles" json:"totalUncles"`
}

type Enode struct {
	Id   enode.ID `json:"id"`
	Ip   net.IP   `json:"ip"`
	Name string   `json:"name"`
	TCP  int      `json:"tcp"`
	UDP  int      `json:"tcp"`
}

type NumberChart struct {
	Name       string   `bson:"name" json:"name"`
	Series     []uint64 `bson:"series" json:"series"`
	Timestamps []string `bson:"timestamps" json:"timestamps"`
}

type NumberStringChart struct {
	Name       string   `bson:"name" json:"name"`
	Series     []string `bson:"series" json:"series"`
	Timestamps []string `bson:"timestamps" json:"timestamps"`
}

type MultiSeriesChart struct {
	Name       string               `bson:"name" json:"name"`
	Datasets   []MultiSeriesDataset `bson:"datasets" json:"datasets"`
	Timestamps []string             `bson:"timestamps" json:"timestamps"`
}

type MultiSeriesDataset struct {
	Name       string   `bson:"name" json:"name"`
	Series     []uint   `bson:"series" json:"series"`
	Timestamps []string `bson:"timestamps" json:"timestamps"`
}
