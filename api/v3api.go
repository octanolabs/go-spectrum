package api

import (
	"github.com/gin-gonic/gin"
	"github.com/octanolabs/go-spectrum/models"
	log "github.com/sirupsen/logrus"
	"github.com/ubiq/go-ubiq/rpc"
)

type V3api interface {
	//blocks
	LatestBlock() (models.Block, error)
	LatestBlocks(limit int) ([]models.Block, error)
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
	TokenTransferCount(hash string) (int64, error)
	TokenTransferCountByContract(hash string) (int64, error)

	//misc
	ChainStore(symbol string) (models.Store, error)
}

func v3RouterHandler(server *rpc.Server) gin.HandlerFunc {
	return func(context *gin.Context) {
		server.ServeHTTP(context.Writer, context.Request)
	}
}

func NewV3ServerStart(backend V3api, cfg *Config) {

	server := rpc.NewServer()

	err := server.RegisterName("explorer", backend)

	if err != nil {
		log.Errorf("Error: couldn't register service: %v", err)
	}

	router := gin.Default()

	_ = router.Group("v3", v3RouterHandler(server))

	go func() {
		err := router.Run(":" + cfg.Port)

		if err != nil {
			log.Fatal("Error: Couldn't serve v3 api: %v", err)
		}
	}()

}
