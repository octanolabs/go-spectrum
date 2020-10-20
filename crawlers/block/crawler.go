package block

import (
	lru "github.com/hashicorp/golang-lru"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/log"
	"math/big"
)

const (
	blockCacheLimit = 10
)

type blockCache struct {
	Supply *big.Int `json:"supply"`
	Hash   string   `json:"hash"`
}

type Config struct {
	Enabled     bool   `json:"enabled"`
	Interval    string `json:"interval"`
	MaxRoutines int    `json:"routines"`
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
