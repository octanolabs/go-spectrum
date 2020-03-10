package crawler

import (
	"math/big"
	"net/url"
	"os"
	"time"

	"github.com/octanolabs/go-spectrum/storage"

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
	Enabled     bool   `json:"enabled"`
	Interval    string `json:"interval"`
	MaxRoutines int    `json:"routines"`
	NodeCrawler bool   `json:"node_crawler"`
}

type BlockCrawler struct {
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

func NewBlockCrawler(db *storage.MongoDB, cfg *Config, logger log.Logger, rpc *rpc.RPCClient) *BlockCrawler {
	bc, _ := lru.New(blockCacheLimit)

	if cfg.NodeCrawler {
		nc := NewNodeCrawler(db, cfg, logger.New("module", "node_crawler"))

		nc.Start()
	}

	return &BlockCrawler{db, rpc, cfg, make(chan *logObject), struct{ syncing, reorg bool }{false, false}, bc, logger}
}

func (c *BlockCrawler) Start() {
	c.logger.Info("Starting block BlockCrawler")

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

	interval, err := time.ParseDuration(c.cfg.Interval)
	if err != nil {
		c.logger.Error("can't parse duration", "err", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(interval)

	c.logger.Info("refresh interval: ", interval)

	go c.SyncLoop()

	go func() {
		for {
			select {
			case <-ticker.C:
				c.logger.Debug("loop iteration", "syncing", c.state.syncing)
				if c.state.syncing != true {
					go c.SyncLoop()
				}
			}
		}
	}()

}
