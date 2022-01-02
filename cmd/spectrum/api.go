package main

import (
	"github.com/octanolabs/go-spectrum/api"
	"github.com/octanolabs/go-spectrum/storage"
	"github.com/ubiq/go-ubiq/v6/log"
)

func startApi(mongo *storage.MongoDB, cfg *api.Config, logger log.Logger) {
	a := api.NewV3ApiServer(mongo, cfg, logger)
	a.Start()
}
