package block

import (
	"math/big"

	lru "github.com/hashicorp/golang-lru"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/v7/log"
)

const (
	blockCacheLimit = 10
)

type blockCache struct {
	Supply      *big.Int `json:"supply"`
	Hash        string   `json:"hash"`
	TotalBurned *big.Int `json:"totalBurned"`
}

type Config struct {
	Enabled     bool   `json:"enabled"`
	Interval    string `json:"interval"`
	MaxRoutines int    `json:"routines"`
	Tracing     struct {
		StartBlock uint64 `json:"start_block"`
		BatchSize  int    `json:"batch_size"`
	} `json:"tracing"`
}

type Crawler struct {
	backend *storage.MongoDB
	rpc     *rpc.RPCClient
	cfg     *Config
	logChan chan *logObject
	state   struct {
		syncing bool
		reorg   bool
	}
	blockCache *lru.Cache // Cache for the most recent blocks
	logger     log.Logger
}

func NewBlockCrawler(db *storage.MongoDB, cfg *Config, logger log.Logger, rpc *rpc.RPCClient) *Crawler {
	bc, _ := lru.New(blockCacheLimit)

	return &Crawler{db, rpc, cfg, make(chan *logObject), struct{ syncing, reorg bool }{false, false}, bc, logger}
}
