package crawler

import (
	"github.com/octanolabs/go-spectrum/crawler/block"
	"github.com/octanolabs/go-spectrum/crawler/database"
	"github.com/ubiq/go-ubiq/log"
	"os"
	"time"
)

type Crawler interface {
	RunLoop()
}

type Config struct {
	Enabled         bool            `json:"enabled"`
	Interval        string          `json:"interval"`
	BlockCrawler    block.Config    `json:"blocks"`
	DatabaseCrawler database.Config `json:"database"`
}

func runCrawlers(crawlers []Crawler) {
	for _, v := range crawlers {
		go v.RunLoop()
	}
}

func RunCrawlers(crawlers []Crawler, cfg *Config, logger log.Logger) {

	interval, err := time.ParseDuration(cfg.Interval)
	if err != nil {
		logger.Error("can't parse duration", "err", err)
		os.Exit(1)
	}

	ticker := time.NewTicker(interval)

	logger.Info("refresh interval set", "d", interval)

	runCrawlers(crawlers)

	go func() {
		for {
			select {
			case <-ticker.C:
				runCrawlers(crawlers)
			}
		}
	}()
}
