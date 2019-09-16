package models

import (
	"github.com/octanolabs/go-spectrum/util"
)

type RawUncle struct {
	Number      string `bson:"number" json:"number"`
	Position    string `bson:"position" json:"position"`
	BlockNumber string `bson:"blockNumber" json:"blockNumber"`
	Hash        string `bson:"hash" json:"hash"`
	ParentHash  string `bson:"parentHash" json:"parentHash"`
	Sha3Uncles  string `bson:"sha3Uncles" json:"sha3Uncles"`
	Miner       string `bson:"miner" json:"miner"`
	Difficulty  string `bson:"difficulty" json:"difficulty"`
	GasUsed     string `bson:"gasUsed" json:"gasUsed"`
	GasLimit    string `bson:"gasLimit" json:"gasLimit"`
	Timestamp   string `bson:"timestamp" json:"timestamp"`
	Reward      string `bson:"reward" json:"reward"`
}

func (rw *RawUncle) Convert() *Uncle {
	return &Uncle{
		Number:     util.DecodeHex(rw.Number),
		Position:   util.DecodeHex(rw.Position),
		Hash:       rw.Hash,
		ParentHash: rw.ParentHash,
		Sha3Uncles: rw.Sha3Uncles,
		Miner:      rw.Miner,
		Difficulty: rw.Difficulty,
		GasUsed:    util.DecodeHex(rw.GasUsed),
		GasLimit:   util.DecodeHex(rw.GasLimit),
		Timestamp:  util.DecodeHex(rw.Timestamp),
	}
}

type Uncle struct {
	Number      uint64 `bson:"number" json:"number"`
	Position    uint64 `bson:"position" json:"position"`
	BlockNumber uint64 `bson:"blockNumber" json:"blockNumber"`
	Hash        string `bson:"hash" json:"hash"`
	ParentHash  string `bson:"parentHash" json:"parentHash"`
	Sha3Uncles  string `bson:"sha3Uncles" json:"sha3Uncles"`
	Miner       string `bson:"miner" json:"miner"`
	Difficulty  string `bson:"difficulty" json:"difficulty"`
	GasUsed     uint64 `bson:"gasUsed" json:"gasUsed"`
	GasLimit    uint64 `bson:"gasLimit" json:"gasLimit"`
	Timestamp   uint64 `bson:"timestamp" json:"timestamp"`
	Reward      string `bson:"reward" json:"reward"`
}
