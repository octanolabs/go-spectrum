package crawler

import (
	"math/big"
	"net/url"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	lru "github.com/hashicorp/golang-lru"
	"github.com/octanolabs/go-spectrum/models"
)

const (
	blockCacheLimit = 10
)

type blockCache struct {
	Supply *big.Int `json:"supply"`
	Hash   string   `json:"hash"`
}

type Config struct {
	Interval    string `json:"interval"`
	MaxRoutines int    `json:"routines"`
}

type RPCClient interface {
	GetLatestBlock() (*models.Block, error)
	GetBlockByHeight(uint64) (*models.Block, error)
	GetBlockByHash(string) (*models.Block, error)
	GetUncleByBlockNumberAndIndex(uint64, int) (*models.Uncle, error)
	LatestBlockNumber() (uint64, error)
	GetTxReceipt(string) (*models.TxReceipt, error)
	GetUnclesInBlock([]string, uint64) []*models.Uncle
	Ping() error
}

type Database interface {
	// Init
	Init()

	// storage
	IsFirstRun() bool
	ChainStore(symbol string) (models.Store, error)
	Ping() error

	// getters

	LatestBlock() (models.Block, error)
	BlockByNumber(height uint64) (*models.Block, error)
	PurgeBlock(height uint64) error

	// setters
	AddTransaction(tx *models.Transaction) error
	AddTokenTransfer(tt *models.TokenTransfer) error
	AddUncle(u *models.Uncle) error
	AddBlock(b *models.Block) error
	AddForkedBlock(b *models.Block) error
}

type Crawler struct {
	backend Database
	rpc     RPCClient
	cfg     *Config
	state   struct {
		syncing bool
		reorg   bool
	}
	blockCache *lru.Cache // Cache for the most recent blocks
}

func New(db Database, rpc RPCClient, cfg *Config) *Crawler {
	bc, _ := lru.New(blockCacheLimit)
	return &Crawler{db, rpc, cfg, struct{ syncing, reorg bool }{false, false}, bc}
}

func (c *Crawler) Start() {
	log.Println("Starting block Crawler")

	err := c.rpc.Ping()

	if err != nil {
		if err == err.(*url.Error) {
			log.Errorf("Gubiq node offline")
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
