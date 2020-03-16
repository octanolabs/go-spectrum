package main

import (
	"github.com/octanolabs/go-spectrum/crawler"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/log"
)

func startCrawler(mongo *storage.MongoDB, cfg *crawler.Config, logger log.Logger, rpc *rpc.RPCClient) {
	c := crawler.NewBlockCrawler(mongo, cfg, logger, rpc)
	c.Start()
}
