package crawler

import (
	"math/big"
	"net/url"
	"os"
	"time"

	"github.com/octanolabs/go-spectrum/storage"

	lru "github.com/hashicorp/golang-lru"
	"github.com/octanolabs/go-spectrum/rpc"
	log "github.com/sirupsen/logrus"
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
}

func New(db *storage.MongoDB, rpc *rpc.RPCClient, cfg *Config) *Crawler {
	bc, _ := lru.New(blockCacheLimit)
	return &Crawler{db, rpc, cfg, make(chan *logObject), struct{ syncing, reorg bool }{false, false}, bc}
}

func (c *Crawler) Start() {
	log.Println("Starting block Crawler")

	err := c.rpc.Ping()

	startLogger(c)

	if err != nil {
		if err == err.(*url.Error) {
			log.Errorf("Gubiq node offline: %v", err)
			os.Exit(1)
		} else {
			log.Errorf("Error pinging rpc node: %#v", err)
		}
	}

	if c.backend.IsFirstRun() {
		c.backend.Init()
	}

	interval, err := time.ParseDuration(c.cfg.Interval)
	if err != nil {
		log.Fatalf("Crawler: can't parse duration: %v", err)
	}

	ticker := time.NewTicker(interval)

	log.Printf("Block refresh interval: %v", interval)

	go c.SyncLoop()

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Debugf("Loop: %v, sync: %v", time.Now().UTC(), c.state.syncing)
				if c.state.syncing != true {
					go c.SyncLoop()
				}
			}
		}
	}()

}
