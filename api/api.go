package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ubiq/go-ubiq/v3/log"
	"github.com/ubiq/go-ubiq/v3/rpc"
)

type Config struct {
	Enabled bool `json:"enabled"`
	//V3      bool   `json:"v4"`
	Host string `json:"host"`
	Port string `json:"port"`
	//Nodemap struct {
	//	Enabled bool   `json:"enabled"`
	//	Mode    string `json:"mode"`
	//	Geodb   string `json:"mmdb"`
	//} `json:"nodemap"`
}

type ApiServer struct {
	handlers v4api
	cfg      *Config
	logger   log.Logger
}

func (a *ApiServer) Start() {

	rpcServer := rpc.NewServer()

	err := rpcServer.RegisterName("explorer", a.handlers)

	if err != nil {
		a.logger.Error("Error: couldn't register service: ", err)
	}

	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Content-Type"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	v3 := router.Group("v3")

	v3.Use(v3ConvertRequest())
	v3.Use(jsonParserMiddleware())
	v3.Use(jsonLoggerMiddleware(a.logger.New("endpoint", "/v3")))
	v3.Use(v3ConvertResponse())

	{
		v3.GET("/*path", v4RouterHandler(rpcServer))
	}

	v4 := router.Group("v4")

	v4.Use(jsonParserMiddleware())
	v4.Use(jsonLoggerMiddleware(a.logger.New("endpoint", "/v4")))

	// Sending a request without and id field will return an empty body
	{
		v4.POST("/", v4RouterHandler(rpcServer))
	}

	go func() {
		err := router.Run(a.cfg.Host + ":" + a.cfg.Port)

		if err != nil {
			log.Error("Couldn't run router", "err", err)
		}
	}()
}

func NewV3ApiServer(backend v4api, cfg *Config, logger log.Logger) *ApiServer {

	s := &ApiServer{
		handlers: backend,
		cfg:      cfg,
		logger:   logger,
	}

	return s
}
