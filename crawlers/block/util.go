package block

import (
	"math/big"
	"time"

	"github.com/ubiq/go-ubiq/v6/consensus/ubqhash"
	"github.com/ubiq/go-ubiq/v6/params"

	"github.com/ubiq/go-ubiq/v6/log"

	"github.com/octanolabs/go-spectrum/models"
)

var (
	config = params.MainnetChainConfig
	big32  = big.NewInt(32)
)

type logObject struct {
	blockNo           uint64
	blocks            int
	txns              int
	tokentransfers    int
	contractsDeployed int
	contractCalls     int
	uncleNo           int
	minted            *big.Int
	supply            *big.Int
}

func (l *logObject) add(o *logObject) {
	l.blockNo = o.blockNo
	l.blocks++
	l.txns += o.txns
	l.tokentransfers += o.tokentransfers
	l.contractsDeployed += o.contractsDeployed
	l.contractCalls += o.contractCalls
	l.uncleNo += o.uncleNo
	l.minted.Add(l.minted, o.minted)
	l.supply = o.supply
}

func (l *logObject) clear() {
	l.blockNo = 0
	l.blocks = 0
	l.txns = 0
	l.tokentransfers = 0
	l.contractsDeployed = 0
	l.contractCalls = 0
	l.uncleNo = 0
	l.minted = new(big.Int)
	l.supply = new(big.Int)
}

func startLogger(c chan *logObject, logger log.Logger) {

	// Start logging goroutine

	go func(ch chan *logObject) {
		start := time.Now()
		stats := &logObject{
			0,
			0,
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
						logger.Info("Imported new chain segment",
							"blocks", stats.blocks,
							"head", stats.blockNo,
							"transactions", stats.txns,
							"transfers", stats.tokentransfers,
							"contracts", stats.contractsDeployed,
							"contractCalls", stats.contractCalls,
							"uncles", stats.uncleNo,
							"minted", stats.minted,
							"supply", stats.supply,
							"took", time.Since(start))

						stats.clear()
						start = time.Now()
					}
				} else {
					if stats.blocks > 0 {
						logger.Info("Imported new chain segment",
							"blocks", stats.blocks,
							"transactions", stats.txns,
							"transfers", stats.tokentransfers,
							"contracts", stats.contractsDeployed,
							"contractCalls", stats.contractCalls,
							"uncles", stats.uncleNo,
							"minted", stats.minted,
							"supply", stats.supply,
							"took", time.Since(start))
					}
					break logLoop
				}
			}
		}
	}(c)
}

func AccumulateRewards(block *models.Block, uncles []models.Uncle) (*big.Int, []*big.Int, *big.Int) {

	var (
		blockNo      = new(big.Int).SetUint64(block.Number)
		minted       = new(big.Int)
		blockReward  = new(big.Int)
		uncleRewards = make([]*big.Int, 0)
	)

	// block reward (miner)
	initialReward, blockReward := ubqhash.CalcBaseBlockReward(config.Ubqhash, blockNo, config.IsLondon(blockNo))

	// Uncle reward step down fix. (activates along-side byzantium)
	// pre-byzantium uncle reward calculation did not take into account monetary policy step-downs,
	// always calculating uncle rewards using biggest possible block reward

	ufixReward := initialReward
	if config.IsByzantium(blockNo) {
		ufixReward = blockReward
	}

	for _, uncle := range uncles {
		uncleNo := new(big.Int).SetUint64(uncle.Number)

		// uncle block miner reward (depth === 1 ? baseBlockReward * 0.5 : 0)
		uncleReward := ubqhash.CalcUncleBlockReward(config, blockNo, uncleNo, ufixReward)

		// add reward for the miner who mined this uncle
		minted.Add(minted, uncleReward)
		uncleRewards = append(uncleRewards, uncleReward)

		// add reward to block miner for including this uncle (baseBlockReward/32)
		bonusReward := new(big.Int).Set(uncleReward).Div(ufixReward, big32)
		blockReward.Add(blockReward, bonusReward)
	}

	// add reward for block miner
	minted.Add(minted, blockReward)

	return blockReward, uncleRewards, minted
}
