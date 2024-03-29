package models

import (
	"math/big"

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
	//TODO: For indexing purposes "number" should be renamed to "blockNumber"?
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
	//TODO: These are not used for anything
	BlockReward   string `bson:"blockReward" json:"blockReward"`
	UnclesRewards string `bson:"unclesReward" json:"unclesReward"`
	AvgGasPrice   string `bson:"avgGasPrice" json:"avgGasPrice"`
	TxFees        string `bson:"txFees" json:"txFees"`
	//
	ExtraData     string `bson:"extraData" json:"extraData"`
	BaseFeePerGas string `bson:"baseFeePerGas" json:"baseFeePerGas,omitempty"`
	Burned        string `bson:"burned" json:"burned,omitempty"`
	TotalBurned   string `bson:"totalBurned" json:"totalBurned,omitempty"`
}

func (b *RawBlock) Convert() Block {
	gasUsed := util.DecodeHex(b.GasUsed)
	baseFeePerGas := util.DecodeValueHex(b.BaseFeePerGas)
	burned := new(big.Int).SetUint64(gasUsed)
	bigBaseFeePerGas, _ := new(big.Int).SetString(baseFeePerGas, 10)
	burned = burned.Mul(bigBaseFeePerGas, burned)
	return Block{
		Number:          util.DecodeHex(b.Number),
		Timestamp:       util.DecodeHex(b.Timestamp),
		Transactions:    make([]Transaction, len(b.Transactions)),
		RawTransactions: b.Transactions,
		Hash:            b.Hash,
		ParentHash:      b.ParentHash,
		Sha3Uncles:      b.Sha3Uncles,
		Miner:           b.Miner,
		Difficulty:      util.DecodeValueHex(b.Difficulty),
		TotalDifficulty: util.DecodeValueHex(b.TotalDifficulty),
		Size:            util.DecodeHex(b.Size),
		GasUsed:         gasUsed,
		GasLimit:        util.DecodeHex(b.GasLimit),
		Nonce:           b.Nonce,
		Uncles:          b.Uncles,
		UncleNo:         len(b.Uncles),
		// Empty
		BlockReward:  "0",
		UncleRewards: "0",
		AvgGasPrice:  "0",
		TxFees:       "0",
		//
		Minted: "0",
		Supply: "0",
		//
		ExtraData:     b.ExtraData,
		BaseFeePerGas: baseFeePerGas,
		Burned:        burned.String(),
		TotalBurned:   "0",
	}
}

type Block struct {
	Number          uint64           `bson:"number" json:"number"`
	Timestamp       uint64           `bson:"timestamp" json:"timestamp"`
	Transactions    []Transaction    `bson:"transactions" json:"transactions"`
	RawTransactions []RawTransaction `bson:"-" json:"-"`
	TokenTransfers  int              `bson:"tokenTransfers" json:"tokenTransfers"`
	Hash            string           `bson:"hash" json:"hash"`
	ParentHash      string           `bson:"parentHash" json:"parentHash"`
	Sha3Uncles      string           `bson:"sha3Uncles" json:"sha3Uncles"`
	Miner           string           `bson:"miner" json:"miner"`
	Difficulty      string           `bson:"difficulty" json:"difficulty"`
	TotalDifficulty string           `bson:"totalDifficulty" json:"totalDifficulty"`
	Size            uint64           `bson:"size" json:"size"`
	GasUsed         uint64           `bson:"gasUsed" json:"gasUsed"`
	GasLimit        uint64           `bson:"gasLimit" json:"gasLimit"`
	Nonce           string           `bson:"nonce" json:"nonce"`
	// Same as Txs
	Uncles  []string `bson:"-" json:"-"`
	UncleNo int      `bson:"uncles" json:"uncles"`
	// TODO: Should these be strings
	// Maybe make this more clear // minted = baseBlockReward + uncleRewards? + bonusRewardForUncles?
	BlockReward  string `bson:"blockReward" json:"blockReward"`
	UncleRewards string `bson:"uncleRewards" json:"uncleRewards"`
	AvgGasPrice  string `bson:"avgGasPrice" json:"avgGasPrice"`
	TxFees       string `bson:"txFees" json:"txFees"`
	// Supply
	Minted string `bson:"minted" json:"minted"`
	Supply string `bson:"supply" json:"supply"`
	//
	ExtraData     string `bson:"extraData" json:"extraData"`
	BaseFeePerGas string `bson:"baseFeePerGas" json:"baseFeePerGas,omitempty"`
	Burned        string `bson:"burned" json:"burned,omitempty"`
	TotalBurned   string `bson:"totalBurned" json:"totalBurned,omitempty"`
	//
	Trace []BlockTrace `bson:"trace" json:"trace,omitempty"`
	//
	ITransactions []ITransaction `bson:"iTransactions" json:"iTransactions,omitempty"`
}

type StructLog struct {
	Depth   uint64   `bson:"depth" json:"depth"`
	Gas     uint64   `bson:"gas" json:"gas"`
	GasCost uint64   `bson:"gasCost" json:"gasCost"`
	Op      string   `bson:"op" json:"op"`
	Pc      uint64   `bson:"pc" json:"pc"`
	Stack   []string `bson:"stack" json:"stack"`
}

type RawBlockTrace struct {
	Result BlockTrace `bson:"result" json:"result,omitempty"`
}

type BlockTrace struct {
	Failed      bool        `bson:"failed" json:"failed"`
	Gas         uint64      `bson:"gas" json:"gas"`
	ReturnValue string      `bson:"returnValue" json:"returnValue"`
	StructLogs  []StructLog `bson:"structLogs" json:"structLogs"`
}

func (rbt *RawBlockTrace) Convert() BlockTrace {
	return BlockTrace{
		Failed:      rbt.Result.Failed,
		Gas:         rbt.Result.Gas,
		ReturnValue: rbt.Result.ReturnValue,
		StructLogs:  rbt.Result.StructLogs,
	}
}
