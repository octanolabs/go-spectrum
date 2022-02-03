package models

import (
	"net"
	"time"

	"github.com/ubiq/go-ubiq/v6/p2p/enode"
)

const (
	BLOCKS        = "blocks"
	TRANSACTIONS  = "transactions"
	ITRANSACTIONS = "itransactions"
	UNCLES        = "uncles"
	CONTRACTS     = "contracts"
	CONTRACTCALLS = "contractcalls"
	TRANSFERS     = "tokentransfers"
	FORKEDBLOCKS  = "forkedblocks"
	CHARTS        = "charts"
	STORE         = "sysstores"
	ENODES        = "enodes"
	ACCOUNTS      = "accounts"
)

type Store struct {
	Symbol string `bson:"symbol" json:"symbol"`

	Timestamp   int64  `bson:"updated" json:"updated"`
	Supply      string `bson:"supply" json:"supply"`
	LatestBlock Block  `bson:"latestBlock" json:"latestBlock"`

	LatestTraceHash string `json:"latestTraceHash" bson:"latestTraceHash"`

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
	UDP  int      `json:"udp"`
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
	Name       string                `bson:"name" json:"name"`
	Datasets   []*MultiSeriesDataset `bson:"datasets" json:"datasets"`
	Timestamps []string              `bson:"timestamps" json:"timestamps"`
}

type MultiSeriesDataset struct {
	Name       string   `bson:"name" json:"name"`
	Series     []uint   `bson:"series" json:"series"`
	Timestamps []string `bson:"timestamps" json:"timestamps"`
}

func (msd MultiSeriesDataset) Len() int {
	return len(msd.Series)
}

func (msd *MultiSeriesDataset) SliceTime(daysBack int) {

	from, to := 0, msd.Len()

	endOfYesterday := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day()-1, 23, 59, 59, 0, time.Local)

	timeLimit := endOfYesterday.Add(time.Duration(-daysBack) * (24 * time.Hour))

	for k, v := range msd.Timestamps {
		ts, _ := time.Parse("01/02/06", v)

		if ts.After(timeLimit) {
			break
		}
		from = k
	}

	msd.Timestamps = msd.Timestamps[from:to]
	msd.Series = msd.Series[from:to]

}
