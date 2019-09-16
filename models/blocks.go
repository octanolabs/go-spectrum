package models

import (
	"github.com/octanolabs/go-spectrum/util"
)

type RawBlockDetails struct {
	Number string `bson:"number" json:"number"`
	Hash   string `bson:"hash" json:"hash"`
}

func (rbn *RawBlockDetails) Convert() (uint64, string) {
	return util.DecodeHex(rbn.Number), rbn.Hash
}

type RawBlock struct {
	Number          string           `bson:"number" json:"number"`
	Timestamp       string           `bson:"timestamp" json:"timestamp"`
	Transactions    []RawTransaction `bson:"transactions" json:"transactions"`
	Hash            string           `bson:"hash" json:"hash"`
	ParentHash      string           `bson:"parentHash" json:"parentHash"`
	Sha3Uncles      string           `bson:"sha3Uncles" json:"sha3Uncles"`
	Miner           string           `bson:"miner" json:"miner"`
	Difficulty      string           `bson:"difficulty" json:"difficulty"`
	TotalDifficulty string           `bson:"totalDifficulty" json:"totalDifficulty"`
	Size            string           `bson:"size" json:"size"`
	GasUsed         string           `bson:"gasUsed" json:"gasUsed"`
	GasLimit        string           `bson:"gasLimit" json:"gasLimit"`
	Nonce           string           `bson:"nonce" json:"nonce"`
	Uncles          []string         `bson:"uncles" json:"uncles"`
	//
	BlockReward  string `bson:"blockReward" json:"blockReward"`
	UnclesReward string `bson:"unclesReward" json:"unclesReward"`
	AvgGasPrice  string `bson:"avgGasPrice" json:"avgGasPrice"`
	TxFees       string `bson:"txFees" json:"txFees"`
	//
	ExtraData string `bson:"extraData" json:"extraData"`
}

func (b *RawBlock) Convert() *Block {
	return &Block{
		Number:          util.DecodeHex(b.Number),
		Timestamp:       util.DecodeHex(b.Timestamp),
		Transactions:    b.Transactions,
		Txs:             len(b.Transactions),
		Hash:            b.Hash,
		ParentHash:      b.ParentHash,
		Sha3Uncles:      b.Sha3Uncles,
		Miner:           b.Miner,
		Difficulty:      util.DecodeValueHex(b.Difficulty),
		TotalDifficulty: util.DecodeValueHex(b.TotalDifficulty),
		Size:            util.DecodeHex(b.Size),
		GasUsed:         util.DecodeHex(b.GasUsed),
		GasLimit:        util.DecodeHex(b.GasLimit),
		Nonce:           b.Nonce,
		Uncles:          b.Uncles,
		UncleNo:         len(b.Uncles),
		// Empty
		BlockReward:  "0",
		UnclesReward: "0",
		AvgGasPrice:  "0",
		TxFees:       "0",
		//
		ExtraData: b.ExtraData,
	}
}

type Block struct {
	Number    uint64 `bson:"number" json:"number"`
	Timestamp uint64 `bson:"timestamp" json:"timestamp"`
	//
	// Transactions contains raw transactions to be processed, is not encoded in db.
	// Txs is the number of txs in a block, is encoded as "transactions"
	//
	Transactions []RawTransaction `bson:"-" json:"-"`
	Txs          int              `bson:"transactions" json:"transactions"`
	//
	Hash            string `bson:"hash" json:"hash"`
	ParentHash      string `bson:"parentHash" json:"parentHash"`
	Sha3Uncles      string `bson:"sha3Uncles" json:"sha3Uncles"`
	Miner           string `bson:"miner" json:"miner"`
	Difficulty      string `bson:"difficulty" json:"difficulty"`
	TotalDifficulty string `bson:"totalDifficulty" json:"totalDifficulty"`
	Size            uint64 `bson:"size" json:"size"`
	GasUsed         uint64 `bson:"gasUsed" json:"gasUsed"`
	GasLimit        uint64 `bson:"gasLimit" json:"gasLimit"`
	Nonce           string `bson:"nonce" json:"nonce"`
	// Same as Txs
	Uncles  []string `bson:"-" json:"-"`
	UncleNo int      `bson:"uncles" json:"uncles"`
	// TODO: These should be strings
	BlockReward  string `bson:"blockReward" json:"blockReward"`
	UnclesReward string `bson:"unclesReward" json:"unclesReward"`
	AvgGasPrice  string `bson:"avgGasPrice" json:"avgGasPrice"`
	TxFees       string `bson:"txFees" json:"txFees"`
	//
	ExtraData string `bson:"extraData" json:"extraData"`
}
