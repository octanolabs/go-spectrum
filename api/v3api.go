package api

import (
	"github.com/octanolabs/go-spectrum/models"
	log "github.com/sirupsen/logrus"
	"github.com/ubiq/go-ubiq/rpc"
	"net"
	"net/http"
)

type V3api interface {
	ChainStore(symbol string) (models.Store, error)
	BlockByNumber(number uint64) (models.Block, error)
	BlockByHash(hash string) (models.Block, error)
	LatestBlock() (models.Block, error)
	LatestBlocks(limit int) ([]models.Block, error)
	TotalBlockCount() (int64, error)
	UncleByHash(hash string) (models.Uncle, error)
	LatestUncles(limit int64) ([]models.Uncle, error)
	TotalUncleCount() (int64, error)
	ForkedBlockByNumber(number uint64) (models.Block, error)
	LatestForkedBlocks(limit int64) ([]models.Block, error)
	TransactionsByBlockNumber(number uint64) ([]models.Transaction, error)
	TransactionByHash(hash string) (models.Transaction, error)
	TransactionByContractAddress(hash string) (models.Transaction, error)
	LatestTransactions(limit int64) ([]models.Transaction, error)
	LatestTransactionsByAccount(hash string) ([]models.Transaction, error)
	TxnCount(hash string) (int64, error)
	TotalTxnCount() (int64, error)
	TokenTransfersByAccount(token string, account string) ([]models.TokenTransfer, error)
	TokenTransfersByAccountCount(token string, account string) (int64, error)
	LatestTokenTransfersByAccount(hash string) ([]models.TokenTransfer, error)
	LatestTransfersOfToken(hash string) ([]models.TokenTransfer, error)
	TokenTransferCount(hash string) (int64, error)
	TokenTransferCountByContract(hash string) (int64, error)
	LatestTokenTransfers(limit int64) ([]models.TokenTransfer, error)
}

func NewV3ServerStart(backend V3api, cfg *Config) {

	server := rpc.NewServer()

	err := server.RegisterName("explorer", backend)

	if err != nil {
		log.Errorf("Error: couldn't register service: %v", err)
	}

	err = http.ListenAndServe(net.JoinHostPort("0.0.0.0", cfg.Port), server)

	if err != nil {
		log.Errorf("Error: Couldn't serve rpc: %v", err)
	}

}
