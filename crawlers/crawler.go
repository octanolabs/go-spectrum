package crawlers

import (
	"github.com/octanolabs/go-spectrum/crawlers/block"
	"github.com/octanolabs/go-spectrum/crawlers/database"
	"github.com/ubiq/go-ubiq/log"
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
			logger.Error("can't parse duration", "err", err)
			os.Exit(1)
		}

		blockTicker := time.NewTicker(blockInterval)

		logger.Info("blockCrawler interval set", "d", blockTicker)

		go runCrawler(blockTicker, bCrawler)
	}

	if dbCrawler, ok := crawlers["database"]; ok {
		databaseInterval, err := time.ParseDuration(cfg.DatabaseCrawler.Interval)
		if err != nil {
			logger.Error("can't parse duration", "err", err)
			os.Exit(1)
		}

		databaseTicker := time.NewTicker(databaseInterval)

		logger.Info("dbCrawler interval set", "d", databaseTicker)

		go runCrawler(databaseTicker, dbCrawler)
	}
}
