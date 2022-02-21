package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"runtime"

	"github.com/rivo/tview"

	"github.com/octanolabs/go-spectrum/util/logui"

	"github.com/ubiq/go-ubiq/v7/log"

	"github.com/octanolabs/go-spectrum/config"
	"github.com/octanolabs/go-spectrum/params"
	"github.com/octanolabs/go-spectrum/rpc"
	"github.com/octanolabs/go-spectrum/storage"
)

var (
	cfg        config.Config
	appLogger  = log.Root()
	mainLogger log.Logger

	RootHandler *log.GlogHandler

	loguiHandler *logui.PassthroughHandler

	enableLogUi    bool
	logLevel       string
	configFileName string
)

const (
	configFlagDefault = "config.json"
	configFlagDesc    = "specify name of config file (should be in working dir)"

	logLevelFlagDefault = "info"
	logLevelFlagDesc    = "set level of logs"
)

func init() {

	flag.StringVar(&configFileName, "c", configFlagDefault, configFlagDesc)
	flag.StringVar(&configFileName, "config", configFlagDefault, configFlagDesc)

	flag.StringVar(&logLevel, "ll", logLevelFlagDefault, logLevelFlagDesc)
	flag.StringVar(&logLevel, "logLevel", logLevelFlagDefault, logLevelFlagDesc)

	flag.BoolVar(&enableLogUi, "logui", false, "Enables logui")

	flag.Parse()

	if enableLogUi {
		ch := make(chan *tview.TextView, 10)
		loguiHandler = logui.NewPassThroughHandler(ch)

		RootHandler = log.NewGlogHandler(loguiHandler)
	} else {
		RootHandler = log.NewGlogHandler(log.StreamHandler(os.Stdout, log.TerminalFormat(true)))
	}

	if logLevel == "debug" || logLevel == "d" || logLevel == "dbg" {
		RootHandler.Verbosity(log.LvlDebug)
	} else if logLevel == "trace" || logLevel == "t" {
		RootHandler.Verbosity(log.LvlTrace)
	} else {
		RootHandler.Verbosity(log.LvlInfo)
	}

	appLogger.SetHandler(RootHandler)

	mainLogger = log.Root().New("pkg", "main")
}

func main() {
	log.Info(fmt.Sprint("go-spectrum ", params.VersionWithMeta, " (", params.VersionWithCommit, ")"))

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
		mainLogger.Error("can't establish connection to mongo", "err", err)
	} else {
		mainLogger.Info("Successfully connected to mongo", "addr", cfg.Mongo.Address)
	}

	err = mongo.Ping()

	if err != nil {
		mainLogger.Error("Can't establish connection to mongo", "err", err)
	} else {
		mainLogger.Info("mongo: PONG")
	}

	rpcClient := rpc.NewRPCClient(&cfg.Rpc)

	version, err := rpcClient.Ping()

	if err != nil {
		switch err.(type) {
		case *url.Error:
			mainLogger.Error("gubiq node offline", "err", err)
			os.Exit(1)
		default:
			mainLogger.Error(fmt.Sprintf("error pinging gubiq node (%T)", err), "err", err)
		}
	}

	mainLogger.Info("connected to gubiq rpc server", "version", version)

	if mongo.IsFirstRun() {
		mongo.Init(rpcClient)
		mainLogger.Warn("mongo: initialized sysStore, genesis, indexes")
	}

	if cfg.Crawlers.Enabled {
		go startCrawlers(mongo, &cfg.Crawlers, appLogger, rpcClient)
	} else if cfg.Api.Enabled {
		go startApi(mongo, &cfg.Api, appLogger.New("pkg", "api"))
	} else {
		mainLogger.Error("No crawlers enabled. exiting.")
		os.Exit(1)
	}

	if enableLogUi {
		lui := logui.NewLogUi(loguiHandler, appLogger.New("pkg", "ui"))
		lui.Start()
	} else {
		quit := make(chan int)
		<-quit
	}
}
