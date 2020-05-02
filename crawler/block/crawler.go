package block

import (
	"github.com/octanolabs/go-spectrum/storage"
	"math/big"
	"net/url"
	"os"

	lru "github.com/hashicorp/golang-lru"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/ubiq/go-ubiq/log"
)

const (
	blockCacheLimit = 10
)

type blockCache struct {
	Supply *big.Int `json:"supply"`
	Hash   string   `json:"hash"`
}

type Config struct {
	Enabled     bool `json:"enabled"`
	MaxRoutines int  `json:"routines"`
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

func (c *Crawler) Start() {
	c.logger.Info("Starting block Crawler")

	err := c.rpc.Ping()

	if err != nil {
		if err == err.(*url.Error) {
			c.logger.Error("Gubiq node offline", "err", err)
			os.Exit(1)
		} else {
			c.logger.Error("Error pinging rpc node", "err", err)
		}
	}

	if c.backend.IsFirstRun() {
		c.backend.Init()
	}

}
