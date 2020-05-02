package database

import (
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/log"
)

type Crawler struct {
	backend *storage.MongoDB
	cfg     *Config
	logger  log.Logger
}

type Config struct {
	Enabled bool `json:"enabled"`
}

func NewDbCrawler(db *storage.MongoDB, cfg *Config, logger log.Logger) *Crawler {
	return &Crawler{db, cfg, logger}
}

func (c *Crawler) Start() {
	c.logger.Info("Started database crawler")
}

func (c *Crawler) RunLoop() {
	c.logger.Info("Looping database crawler")
}
