package main

import (
	"github.com/octanolabs/go-spectrum/crawlers"
	"github.com/octanolabs/go-spectrum/crawlers/block"
	"github.com/octanolabs/go-spectrum/crawlers/database"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/v6/log"
)

func startCrawlers(mongo *storage.MongoDB, cfg *crawlers.Config, logger log.Logger, rpc *rpc.RPCClient) {

	var crawlerMap = make(map[string]crawlers.Crawler, 2)

	if cfg.BlockCrawler.Enabled {
		blockCrawler := block.NewBlockCrawler(mongo, &cfg.BlockCrawler, logger.New("crawler", "block"), rpc)
		logger.Info("Starting block Crawler")
		crawlerMap["blocks"] = blockCrawler
	}

	if cfg.DatabaseCrawler.Enabled {
		dbCrawler := database.NewDbCrawler(mongo, &cfg.DatabaseCrawler, logger.New("crawler", "database"))
		logger.Info("Starting database crawler")
		crawlerMap["database"] = dbCrawler

	}

	crawlers.RunCrawlers(crawlerMap, cfg, logger)
}
