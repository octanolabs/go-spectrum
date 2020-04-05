package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/octanolabs/go-spectrum/models"
	"github.com/ubiq/go-ubiq/log"
	"github.com/ubiq/go-ubiq/rpc"
)

// use unexported interface so it trims some methods we don't want to serve
type v3api interface {
	//blocks
	LatestBlock() (models.Block, error)
	LatestBlocks(limit int64) ([]models.Block, error)
	BlockByHash(hash string) (models.Block, error)
	BlockByNumber(number uint64) (models.Block, error)
	TransactionsByBlockNumber(number uint64) ([]models.Transaction, error)
	TotalBlockCount() (int64, error)

	//uncles
	LatestUncles(limit int64) ([]models.Uncle, error)
	UncleByHash(hash string) (models.Uncle, error)
	TotalUncleCount() (int64, error)

	//reorgs
	ForkedBlockByNumber(number uint64) (models.Block, error)
	LatestForkedBlocks(limit int64) ([]models.Block, error)

	//txs
	LatestTransactions(limit int64) ([]models.Transaction, error)
	TransactionByHash(hash string) (models.Transaction, error)
	LatestTransactionsByAccount(hash string) ([]models.Transaction, error)
	TransactionByContractAddress(hash string) (models.Transaction, error)
	TxnCount(hash string) (int64, error)
	TotalTxnCount() (int64, error)

	//transfers
	LatestTokenTransfers(limit int64) ([]models.TokenTransfer, error)
	LatestTokenTransfersByAccount(account string) ([]models.TokenTransfer, error)
	TokenTransfersByAccount(token string, account string) ([]models.TokenTransfer, error)
	TokenTransfersByAccountCount(token string, account string) (int64, error)
	LatestTransfersOfToken(hash string) ([]models.TokenTransfer, error)
	ContractTransferCount(hash string) (int64, error)
	TotalTransferCount() (int64, error)
	TokenTransferCountByContract(hash string) (int64, error)

	//misc
	Status() (models.Store, error)
}

func v3RouterHandler(server *rpc.Server) gin.HandlerFunc {
	return func(context *gin.Context) {

		//TODO: added this for dev. Maybe remove in production
		context.Request.Header.Set("Access-Control-Allow-Origin", "localhost:8080")

		server.ServeHTTP(context.Writer, context.Request)
	}
}

func jsonParserMiddleware() gin.HandlerFunc {
	return func(context *gin.Context) {
		method, params, newReader := ParseJsonRequest(context.Request)

		context.Request.Body = newReader
		context.Set("method", method)
		context.Set("params", params)
	}
}

//func jsonLoggerMiddleware() gin.HandlerFunc {
//	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
//
//		//your custom format
//		return fmt.Sprintf("%s - [%d][%s] \t %s | \t %s - %s | %s | %s\t-\t%s\n",
//			param.TimeStamp.Format(time.RFC1123),
//			param.StatusCode,
//			param.Method,
//			param.Latency,
//			param.ClientIP,
//			param.Request.UserAgent(),
//			param.Path,
//			param.Keys["method"],
//			param.Keys["params"],
//		)
//	})
//}

func jsonLoggerMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		context.Next()

		if _, ok := context.Params.Get("method"); ok {

			if _, ok = context.Params.Get("params"); ok {
				logger.Info("received http request",
					"path", context.Request.URL.Path,
					"status", context.Writer.Status(),
					"method", context.Request.Method,
					"latency", time.Since(start),
					"from", context.Request.RemoteAddr,
					"agent", context.Request.UserAgent(),
					"rpcMethod", context.Param("method"),
					"rpcParams", context.Param("params"))
			}
		} else {
			logger.Info("received http request",
				"path", context.Request.URL.Path,
				"status", context.Writer.Status(),
				"method", context.Request.Method,
				"latency", time.Since(start),
				"from", context.Request.RemoteAddr,
				"agent", context.Request.UserAgent())
		}

	}
}
