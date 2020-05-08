package models

import (
	"net"

	"github.com/ubiq/go-ubiq/p2p/enode"
)

const (
	BLOCKS       = "blocks"
	TXNS         = "transactions"
	UNCLES       = "uncles"
	TRANSFERS    = "tokentransfers"
	FORKEDBLOCKS = "forkedblocks"
	CHARTS       = "charts"
	STORE        = "sysstores"
	ENODES       = "enodes"
)

type Store struct {
	Timestamp   int64  `bson:"timestamp" json:"timestamp"`
	Symbol      string `bson:"symbol" json:"symbol"`
	Supply      string `bson:"supply" json:"supply"`
	LatestBlock Block  `bson:"latestBlock" json:"latestBlock"`
	Price       string `bson:"price" json:"price"`

	TotalTransactions   int64 `bson:"totalTransactions" json:"totalTransactions"`
	TotalTokenTransfers int64 `bson:"totalTokenTransfers" json:"totalTokenTransfers"`
	TotalForkedBlocks   int64 `bson:"totalForkedBlocks" json:"totalForkedBlocks"`
	TotalUncles         int64 `bson:"totalUncles" json:"totalUncles"`
}

type Enode struct {
	Id   enode.ID `json:"id"`
	Ip   net.IP   `json:"ip"`
	Name string   `json:"name"`
	TCP  int      `json:"tcp"`
	UDP  int      `json:"tcp"`
}

type Chart struct {
	Name       string      `bson:"name" json:"name"`
	Series     interface{} `bson:"series" json:"series"`
	Timestamps []string    `bson:"timestamps" json:"timestamps"`
}
