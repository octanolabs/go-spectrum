package crawler

import (
	"math/big"

	"github.com/octanolabs/go-spectrum/models"
)

var (
	blockReward *big.Int = big.NewInt(8e+18) // Block reward in wei for successfully mining a block
	big2                 = big.NewInt(2)
	big32                = big.NewInt(32)
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

// AccumulateRewards calculates the mining reward of the given block.
// The total reward consists of the static block reward and rewards for
// included uncles. The total rewards of each uncle block is also returned.
// based on accumulateRewards from gubiq 2.2.0
func AccumulateRewards(block *models.Block, uncles []models.Uncle) (*big.Int, *big.Int, *big.Int) {
	reward := new(big.Int).Set(blockReward)
	blocknum := new(big.Int).SetUint64(block.Number)

	if blocknum.Cmp(big.NewInt(358363)) > 0 {
		reward = big.NewInt(7e+18)
	}
	if blocknum.Cmp(big.NewInt(716727)) > 0 {
		reward = big.NewInt(6e+18)
	}
	if blocknum.Cmp(big.NewInt(1075090)) > 0 {
		reward = big.NewInt(5e+18)
	}
	if blocknum.Cmp(big.NewInt(1433454)) > 0 {
		reward = big.NewInt(4e+18)
	}
	if blocknum.Cmp(big.NewInt(1791818)) > 0 {
		reward = big.NewInt(3e+18)
	}
	if blocknum.Cmp(big.NewInt(2150181)) > 0 {
		reward = big.NewInt(2e+18)
	}
	if blocknum.Cmp(big.NewInt(2508545)) > 0 {
		reward = big.NewInt(1e+18)
	}

	r := new(big.Int)
	u := new(big.Int)
	for _, uncle := range uncles {
		unclenum := new(big.Int).SetUint64(uncle.Number)
		r.Add(unclenum, big2)
		r.Sub(r, blocknum)
		r.Mul(r, blockReward)
		r.Div(r, big2)

		if blocknum.Cmp(big.NewInt(10)) < 0 {
			u.Add(u, r)
			r.Div(blockReward, big32)
			if r.Cmp(big.NewInt(0)) < 0 {
				r = big.NewInt(0)
			}
		} else {
			if r.Cmp(big.NewInt(0)) < 0 {
				r = big.NewInt(0)
			}
			u.Add(u, r)
			r.Div(blockReward, big32)
		}

		reward.Add(reward, r)
	}

	minted := new(big.Int)
	minted.Add(reward, u)
	return reward, u, minted
}
