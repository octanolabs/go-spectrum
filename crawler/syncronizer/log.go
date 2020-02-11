package syncronizer

import (
	"math/big"
	"time"

	log "github.com/sirupsen/logrus"
)

type logObject struct {
	blockNo        uint64
	blocks         int
	txns           int
	tokentransfers int
	uncleNo        int
	minted         *big.Int
	supply         *big.Int
}

func (l *logObject) add(o *logObject) {
	l.blockNo = o.blockNo
	l.blocks++
	l.txns += o.txns
	l.tokentransfers += o.tokentransfers
	l.uncleNo += o.uncleNo
	l.minted.Add(l.minted, o.minted)
	l.supply = o.supply
}

func (l *logObject) clear() {
	l.blockNo = 0
	l.blocks = 0
	l.txns = 0
	l.tokentransfers = 0
	l.uncleNo = 0
	l.minted = new(big.Int)
	l.supply = new(big.Int)
}

func startLogger(s *Synchronizer) {
	lc := make(chan *logObject)

	s.logChan = lc

	// Start logging goroutine

	go func(ch chan *logObject) {
		start := time.Now()
		stats := &logObject{
			0,
			0,
			0,
			0,
			0,
			new(big.Int),
			new(big.Int),
		}
	logLoop:
		for {
			select {
			case lo, more := <-ch:
				if more {
					stats.add(lo)

					if stats.blocks >= 1000 || time.Now().After(start.Add(time.Minute)) {
						log.WithFields(log.Fields{
							"blocks":       stats.blocks,
							"head":         stats.blockNo,
							"transactions": stats.txns,
							"transfers":    stats.tokentransfers,
							"uncles":       stats.uncleNo,
							"minted":       stats.minted,
							"supply":       stats.supply,
							"t":            time.Since(start),
						}).Info("Imported new chain segment")

						stats.clear()
						start = time.Now()
					}
				} else {
					if stats.blocks > 0 {
						log.WithFields(log.Fields{
							"blocks":       stats.blocks,
							"head":         stats.blockNo,
							"transactions": stats.txns,
							"transfers":    stats.tokentransfers,
							"uncles":       stats.uncleNo,
							"minted":       stats.minted,
							"supply":       stats.supply,
							"t":            time.Since(start),
						}).Info("Imported new chain segment")
					}
					break logLoop
				}
			}
		}
	}(lc)
}
