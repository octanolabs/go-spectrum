package config

import (
	"github.com/octanolabs/go-spectrum/api"
	"github.com/octanolabs/go-spectrum/crawlers"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
)

type Config struct {
	Threads  int             `json:"threads"`
	Crawlers crawlers.Config `json:"crawlers"`
	Mongo    storage.Config  `json:"mongo"`
	Rpc      rpc.Config      `json:"rpc"`
	Api      api.Config      `json:"api"`
}

// {
//   "mongodb": {
//     "user": "spectrum",
//     "password": "UBQ4Lyfe",
//     "database": "spectrumdb",
//     "address": "localhost",
//     "port": 27017
//   },
//
//   "mongodbtest": {
//     "user": "spectrum",
//     "password": "UBQ4Lyfe",
//     "database": "spectrum-test",
//     "address": "localhost",
//     "port": 27017
//   }
// }
