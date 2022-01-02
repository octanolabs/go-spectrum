package crawlers

import (
	"github.com/octanolabs/go-spectrum/crawlers/block"
	"github.com/octanolabs/go-spectrum/crawlers/database"
	"github.com/ubiq/go-ubiq/v6/log"
	"os"
	"time"
)

type Crawler interface {
	RunLoop()
}

type Config struct {
	Enabled         bool            `json:"enabled"`
	BlockCrawler    block.Config    `json:"blocks"`
	DatabaseCrawler database.Config `json:"database"`
}

func runCrawler(ticker *time.Ticker, c Crawler) {
	c.RunLoop()
	for {
		select {
		case <-ticker.C:
			c.RunLoop()
		}
	}
}

func RunCrawlers(crawlers map[string]Crawler, cfg *Config, logger log.Logger) {

	if bCrawler, ok := crawlers["blocks"]; ok {
		blockInterval, err := time.ParseDuration(cfg.BlockCrawler.Interval)
		if err != nil {
			logger.Error("can't parse blockCrawler duration", "d", cfg.BlockCrawler.Interval, "err", err)
			os.Exit(1)
		}

		blockTicker := time.NewTicker(blockInterval)

		logger.Warn("blockCrawler interval set", "d", cfg.BlockCrawler.Interval)

		go runCrawler(blockTicker, bCrawler)
	}

	if dbCrawler, ok := crawlers["database"]; ok {
		databaseInterval, err := time.ParseDuration(cfg.DatabaseCrawler.Interval)
		if err != nil {
			logger.Error("can't parse dbCrawler duration", "d", cfg.DatabaseCrawler.Interval, "err", err)
			os.Exit(1)
		}

		databaseTicker := time.NewTicker(databaseInterval)

		logger.Warn("dbCrawler interval set", "d", cfg.DatabaseCrawler.Interval)

		go runCrawler(databaseTicker, dbCrawler)
	}
}
