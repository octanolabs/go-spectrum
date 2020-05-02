package main

import (
	"github.com/octanolabs/go-spectrum/crawler"
	"github.com/octanolabs/go-spectrum/crawler/block"
	"github.com/octanolabs/go-spectrum/crawler/database"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/log"
)

func startCrawlers(mongo *storage.MongoDB, cfg *crawler.Config, logger log.Logger, rpc *rpc.RPCClient) {

	var crawlers = make([]crawler.Crawler, 0)

	if cfg.BlockCrawler.Enabled {
		blockCrawler := block.NewBlockCrawler(mongo, &cfg.BlockCrawler, logger.New("crawler", "block"), rpc)
		blockCrawler.Start()
		crawlers = append(crawlers, blockCrawler)
	}

	if cfg.DatabaseCrawler.Enabled {
		dbCrawler := database.NewDbCrawler(mongo, &cfg.DatabaseCrawler, logger.New("crawler", "database"))
		dbCrawler.Start()
		crawlers = append(crawlers, dbCrawler)
	}

	crawler.RunCrawlers(crawlers, cfg, logger)
}
