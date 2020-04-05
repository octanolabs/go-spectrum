package crawler

import (
	"log"

	"github.com/octanolabs/go-spectrum/storage"
)

type DbCrawler struct {
	backend *storage.MongoDB
	logger  log.Logger
}

func NewDbCrawler(db *storage.MongoDB, logger log.Logger) *DbCrawler {
	return &DbCrawler{db, logger}
}

func (db *DbCrawler) Start() {

}
