package database

import (
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/v6/log"
	"time"
)

type Crawler struct {
	backend *storage.MongoDB
	cfg     *Config
	logger  log.Logger
}

type Config struct {
	Enabled  bool   `json:"enabled"`
	Interval string `json:"interval"`
}

func NewDbCrawler(db *storage.MongoDB, cfg *Config, logger log.Logger) *Crawler {
	return &Crawler{db, cfg, logger}
}

func (c *Crawler) RunLoop() {
	if s, err := c.backend.Status(); err == nil && s.LatestBlock.Number == 0 {
		c.logger.Error("skipping cycle, the database is empty")
	} else {
		start := time.Now()
		c.CrawlBlocks()

		c.logger.Info("crawled blocks collection", "took", time.Since(start))

		start = time.Now()
		c.CrawlTransactions()

		c.logger.Info("crawled transactions collection", "took", time.Since(start))
	}
}

//func (c *Crawler) CrawlTransactions() {
//
//	cursor, err := c.backend.IterTransactions()
//	if err != nil {
//		c.logger.Error("Error creating iter", "err", err)
//	}
//
//	c.handleCursor(cursor, func(cursor *mongo.Cursor, task *syncronizer.Task) {
//		var block models.Block
//		if err := cursor.Decode(&block); err != nil {
//			c.logger.Error("Error decoding bloc", "err", err)
//		}
//	})
//
//}
//
//func (c *Crawler) CrawlTokenTransfers() {
//
//	cursor, err := c.backend.IterTokenTransfers()
//	if err != nil {
//		c.logger.Error("Error creating iter", "err", err)
//	}
//
//	c.handleCursor(cursor, func(cursor *mongo.Cursor, task *syncronizer.Task) {
//		var block models.Block
//		if err := cursor.Decode(&block); err != nil {
//			c.logger.Error("Error decoding bloc", "err", err)
//		}
//	})
//
//}
//
//func (c *Crawler) CrawlUncles() {
//
//	cursor, err := c.backend.IterUncles()
//	if err != nil {
//		c.logger.Error("Error creating iter", "err", err)
//	}
//
//	c.handleCursor(cursor, func(cursor *mongo.Cursor, task *syncronizer.Task) {
//		var block models.Block
//		if err := cursor.Decode(&block); err != nil {
//			c.logger.Error("Error decoding bloc", "err", err)
//		}
//	})
//
//}
//
//func (c *Crawler) CrawlForkedBlocks() {
//
//	cursor, err := c.backend.IterForkedBlocks()
//	if err != nil {
//		c.logger.Error("Error creating iter", "err", err)
//
//		c.handleCursor(cursor, func(cursor *mongo.Cursor, task *syncronizer.Task) {
//			var block models.Block
//			if err := cursor.Decode(&block); err != nil {
//				c.logger.Error("Error decoding bloc", "err", err)
//			}
//		})
//	}
//
//}
