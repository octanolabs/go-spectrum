package api

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/octanolabs/go-spectrum/models"
	"github.com/ubiq/go-ubiq/log"
	"github.com/ubiq/go-ubiq/rpc"
)

// use unexported interface so it trims some methods we don't want to serve
type v4api interface {
	//blocks
	LatestBlock() (models.Block, error)
	BlockByHash(hash string) (models.Block, error)
	BlockByNumber(number uint64) (models.Block, error)
	TransactionsByBlockNumber(number uint64) ([]models.Transaction, error)
	TotalBlockCount() (int64, error)

	//uncles
	UncleByHash(hash string) (models.Uncle, error)
	TotalUncleCount() (int64, error)

	//reorgs
	ForkedBlockByNumber(number uint64) (models.Block, error)
	TotalForkedBlockCount() (int64, error)

	//txs
	TransactionByHash(hash string) (models.Transaction, error)
	TransactionByContractAddress(hash string) (models.Transaction, error)
	TxnCount(hash string) (int64, error)
	TotalTxnCount() (int64, error)

	//transfers
	TokenTransfersByAccount(account string) ([]models.TokenTransfer, error)
	TokenTransfersByAccountCount(account string) (int64, error)
	TransfersOfTokenByAccount(token string, account string) ([]models.TokenTransfer, error)
	TransfersOfTokenByAccountCount(token string, account string) (int64, error)
	TransfersByContract(address string) ([]models.TokenTransfer, error)
	ContractTransferCount(address string) (int64, error)
	TotalTransferCount() (int64, error)

	//charts
	GetNumberChart(name string, limit int) (models.NumberChart, error)
	GetNumberStringChart(name string, limit int) (models.NumberStringChart, error)
	GetMultiSeriesChart(name string, limit int) (models.MultiSeriesChart, error)
	ListCharts() ([]string, error)

	//api-specific
	LatestBlocks(limit int64) (map[string]interface{}, error)
	LatestMinedBlocks(account string, limit int64) (map[string]interface{}, error)
	LatestUncles(limit int64) (map[string]interface{}, error)
	LatestForkedBlocks(limit int64) (map[string]interface{}, error)
	LatestTransactions(limit int64) (map[string]interface{}, error)
	LatestTokenTransfers(limit int64) (map[string]interface{}, error)
	LatestTransfersOfToken(account string) (map[string]interface{}, error)
	LatestTokenTransfersByAccount(account string) (map[string]interface{}, error)
	LatestTransactionsByAccount(account string) (map[string]interface{}, error)
	LatestFailedTransactions(limit int64) (map[string]interface{}, error)
	LatestContractCalls(limit int64) (map[string]interface{}, error)
	LatestContractsDeployed(limit int64) (map[string]interface{}, error)

	//misc
	Status() (models.Store, error)
}

func v4RouterHandler(server *rpc.Server) gin.HandlerFunc {
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

//func convertRawMethod(m interface{}) (result string) {
//	str := fmt.Sprintf("%s", m)
//	split := strings.Split(str, "_")
//	return split[1]
//}
//
//func convertRawParams(p interface{}) (result []string) {
//	result = make([]string, 0)
//	params, ok := p.([]json.RawMessage)
//	if !ok {
//		return
//	}
//	for _, v := range params {
//		result = append(result, fmt.Sprintf("%s", v))
//	}
//	return
//}

func jsonLoggerMiddleware(logger log.Logger) gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()

		context.Next()

		if method, ok := context.Keys["method"]; ok {

			if params, ok := context.Keys["params"]; ok {

				logger.Info("received http request",
					"status", context.Writer.Status(),
					"method", context.Request.Method,
					"latency", time.Since(start),
					"from", context.Request.RemoteAddr,
					"agent", context.Request.UserAgent(),
					"rpcMethod", fmt.Sprintf("%s", method),
					"rpcParams", fmt.Sprintf("%s", params))
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
