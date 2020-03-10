package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/octanolabs/go-spectrum/api"

	"github.com/ubiq/go-ubiq/log"

	"github.com/octanolabs/go-spectrum/config"
	"github.com/octanolabs/go-spectrum/crawler"
	"github.com/octanolabs/go-spectrum/params"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
)

var cfg config.Config

var appLogger log.Logger
var mainLogger log.Logger

func init() {

	glogHandler := log.NewGlogHandler(log.StreamHandler(os.Stdout, log.LogfmtFormat()))

	v, _ := strconv.ParseBool(os.Getenv("DEBUG"))
	if v {
		glogHandler.Verbosity(log.LvlDebug)
	} else {
		glogHandler.Verbosity(log.LvlInfo)
	}

	appLogger = log.New(glogHandler)
	mainLogger = appLogger.New(log.Must.FileHandler("main", log.LogfmtFormat()))
}

func readConfig(cfg *config.Config) {

	if len(os.Args) == 1 {
		mainLogger.Error("Invalid arguments", os.Args)
	}

	conf := os.Args[1]
	confPath, err := filepath.Abs(conf)

	if err != nil {
		mainLogger.Error("Error: could not parse config filepath", "err", err)
	}

	mainLogger.Info("Loading config", "path", confPath)

	configFile, err := os.Open(confPath)
	if err != nil {
		appLogger.Error("File error", "err", err.Error())
	}

	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	if err := jsonParser.Decode(&cfg); err != nil {
		mainLogger.Error("Config error", "err", err.Error())
	}
}

func startCrawler(mongo *storage.MongoDB, cfg *crawler.Config, logger log.Logger, rpc *rpc.RPCClient) {
	c := crawler.NewBlockCrawler(mongo, cfg, logger, rpc)
	c.Start()
}

func startApi(mongo *storage.MongoDB, cfg *api.Config, logger log.Logger) {
	a := api.NewV3ApiServer(mongo, cfg, logger)
	a.Start()
}

func main() {
	log.Info("go-spectrum ", params.VersionWithMeta, " (", params.VersionWithCommit, ")")

	readConfig(&cfg)

	mainLogger.Debug("Printing config", "cfg", cfg)

	if cfg.Threads > 0 {
		runtime.GOMAXPROCS(cfg.Threads)
		mainLogger.Info("App running", "threads", cfg.Threads)
	} else {
		runtime.GOMAXPROCS(1)
		mainLogger.Info("App running with 1 thread")
	}

	mainLogger.Debug("Connecting to mongo", "addr", cfg.Mongo.ConnectionString())

	mongo, err := storage.NewConnection(&cfg.Mongo) // TODO - iquidus: fix this check

	if err != nil {
		mainLogger.Error("Can't establish connection to mongo: %v", err)
	} else {
		mainLogger.Info("Successfully connected to mongo at %v", cfg.Mongo.Address)
	}

	err = mongo.Ping()

	if err != nil {
		mainLogger.Error("Can't establish connection to mongo", "err", err)
	} else {
		mainLogger.Info("mongo: PONG")
	}

	rpcClient := rpc.NewRPCClient(&cfg.Rpc)

	if cfg.Crawler.Enabled {
		go startCrawler(mongo, &cfg.Crawler, appLogger.New("module", "crawler"), rpcClient)
	} else if cfg.Api.Enabled {
		go startApi(mongo, &cfg.Api, appLogger.New("module", "api"))
	}

	quit := make(chan bool)
	<-quit
}
