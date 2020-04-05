package api

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/ubiq/go-ubiq/log"
	"github.com/ubiq/go-ubiq/rpc"
)

type Config struct {
	Enabled bool `json:"enabled"`
	//V3      bool   `json:"v3"`
	Port string `json:"port"`
	//Nodemap struct {
	//	Enabled bool   `json:"enabled"`
	//	Mode    string `json:"mode"`
	//	Geodb   string `json:"mmdb"`
	//} `json:"nodemap"`
}

type ApiServer struct {
	handlers v3api
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
		AllowHeaders:  []string{"Origin"},
		ExposeHeaders: []string{"Content-Length"},
		MaxAge:        12 * time.Hour,
	}))

	v2 := router.Group("v2")

	v2.Use(v2ConvertRequest())
	v2.Use(jsonParserMiddleware())
	v2.Use(jsonLoggerMiddleware(a.logger.New("endpoint", "/v2")))
	v2.Use(v2ConvertResponse(a.handlers))

	{
		v2.GET("/*path", v3RouterHandler(rpcServer))
	}

	v3 := router.Group("v3")

	v3.Use(jsonParserMiddleware())
	v3.Use(jsonLoggerMiddleware(a.logger.New("endpoint", "/v3")))

	{
		v3.POST("/", v3RouterHandler(rpcServer))
	}

	go func() {
		err := router.Run(":" + a.cfg.Port)

		if err != nil {
			log.Error("Couldn't run router", "err", err)
		}
	}()
}

func NewV3ApiServer(backend v3api, cfg *Config, logger log.Logger) *ApiServer {

	s := &ApiServer{
		handlers: backend,
		cfg:      cfg,
		logger:   logger,
	}

	return s
}
