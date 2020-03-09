package models

import (
	"net"

	"github.com/ubiq/go-ubiq/p2p/enode"
)

const (
	BLOCKS    = "blocks"
	TXNS      = "transactions"
	UNCLES    = "uncles"
	TRANSFERS = "tokentransfers"
	REORGS    = "forkedblocks"
	CHARTS    = "charts"
	STORE     = "sysstores"
	ENODES    = "enodes"
)

type Store struct {
	Timestamp int64  `bson:"timestamp" json:"timestamp"`
	Symbol    string `bson:"symbol" json:"symbol"`
	Supply    string `bson:"supply" json:"supply"`
	// LatestBlock Block     `bson:"latestBlock" json:"latestBlock"`
	// Price       string    `bson:"price" json:"price"`
	// syncBlock [1]uint64 `bson:"sync"`
}

type Enode struct {
	Id   enode.ID `json:"id"`
	Ip   net.IP   `json:"ip"`
	Name string   `json:"name"`
	TCP  int      `json:"tcp"`
	UDP  int      `json:"tcp"`
}
