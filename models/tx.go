package models

import (
	"github.com/octanolabs/go-spectrum/util"
	log "github.com/sirupsen/logrus"
)

type RawTransaction struct {
	BlockHash        string `json:"blockHash"`
	BlockNumber      string `json:"blockNumber"`
	From             string `json:"from"`
	Gas              string `json:"gas"`
	GasPrice         string `json:"gasPrice"`
	Hash             string `json:"hash"`
	Input            string `json:"input"`
	Nonce            string `json:"nonce"`
	To               string `json:"to"`
	TransactionIndex string `json:"transactionIndex"`
	Value            string `json:"value"`
	V                string `json:"v"`
	R                string `json:"r"`
	S                string `json:"s"`
}

func (rt *RawTransaction) Convert() Transaction {
	return Transaction{
		BlockHash:   rt.BlockHash,
		BlockNumber: util.DecodeHex(rt.BlockNumber),
		Hash:        rt.Hash,
		//
		// Timestamp
		//
		Input:            rt.Input,
		Value:            util.DecodeValueHex(rt.Value),
		Gas:              util.DecodeHex(rt.Gas),
		GasPrice:         util.DecodeHex(rt.GasPrice),
		Nonce:            rt.Nonce,
		TransactionIndex: util.DecodeHex(rt.TransactionIndex),
		From:             rt.From,
		To:               rt.To,
		//
		// GasUsed         :
		// ContractAddress :
		// Logs            :
		//
	}
}

type Transaction struct {
	BlockHash        string `bson:"blockHash" json:"blockHash"`
	BlockNumber      uint64 `bson:"blockNumber" json:"blockNumber"`
	Hash             string `bson:"hash" json:"hash"`
	Timestamp        uint64 `bson:"timestamp" json:"timestamp"`
	Input            string `bson:"input" json:"input"`
	Value            string `bson:"value" json:"value"`
	Gas              uint64 `bson:"gas" json:"gas"`
	GasPrice         uint64 `bson:"gasPrice" json:"gasPrice"`
	Nonce            string `bson:"nonce" json:"nonce"`
	TransactionIndex uint64 `bson:"transactionIndex" json:"transactionIndex"`
	From             string `bson:"from" json:"from"`
	To               string `bson:"to" json:"to"`
	//
	GasUsed         uint64  `bson:"gasUsed" json:"gasUsed"`
	ContractAddress string  `bson:"contractAddress" json:"contractAddress"`
	Logs            []TxLog `bson:"logs" json:"logs"`
	//
}

func (tx *Transaction) IsTokenTransfer() bool {

	if tx.Input == "0x" || tx.Input == "0x00" {
		return false
	}

	if len(tx.Input) < 10 {
		return false
	}

	switch tx.Input[:10] {
	case "0xa9059cbb":
		return true
	case "0x23b872dd":
		return true
	case "0x6ea056a9":
		return true
	case "0x40c10f19":
		return true
	default:
		return false
	}
}

func (tx *Transaction) GetTokenTransfer() *TokenTransfer {
	var params []string

	method := tx.Input[:10]

	if len(tx.Input) == 138 {
		params = []string{
			tx.Input[10:74], tx.Input[74:],
		}
	} else if len(tx.Input) == 202 {
		params = []string{
			tx.Input[10:74], tx.Input[74:138], tx.Input[138:],
		}
	} else {
		log.Errorf("Error processing token transfer: input length is not standard: len: %v", len(tx.Input))
		return &TokenTransfer{}
	}

	transfer := &TokenTransfer{
		BlockNumber: tx.BlockNumber,
		Hash:        tx.Hash,
		Timestamp:   tx.Timestamp,
	}

	switch method {
	case "0xa9059cbb": // transfer
		transfer.From = tx.From
		transfer.To = util.InputParamsToAddress(params[0])
		transfer.Value = util.DecodeValueHex(params[1])
		transfer.Contract = tx.To
		transfer.Method = "transfer"

		return transfer

	case "0x23b872dd": // transferFrom
		transfer.From = util.InputParamsToAddress(params[0])
		transfer.To = util.InputParamsToAddress(params[1])
		transfer.Value = util.DecodeValueHex(params[2])
		transfer.Contract = tx.To
		transfer.Method = "transferFrom"

		return transfer

	case "0x6ea056a9": // sweep
		transfer.From = tx.To
		transfer.To = tx.From
		transfer.Value = util.DecodeValueHex(params[1])
		transfer.Contract = util.InputParamsToAddress(params[0])
		transfer.Method = "sweep"

		return transfer

	case "0x40c10f19": // mint
		transfer.From = "0x0000000000000000000000000000000000000000"
		transfer.To = util.InputParamsToAddress(params[0])
		transfer.Value = util.DecodeValueHex(params[1])
		transfer.Contract = tx.To
		transfer.Method = "mint"

		return transfer
	default:
		transfer.Method = "unknown"
		transfer.Data = tx.Input

		return transfer
	}

}

type TokenTransfer struct {
	BlockNumber uint64 `bson:"blockNumber" json:"blockNumber"`
	Hash        string `bson:"hash" json:"hash"`
	Timestamp   uint64 `bson:"timestamp" json:"timestamp"`
	From        string `bson:"from" json:"from"`
	To          string `bson:"to" json:"to"`
	Value       string `bson:"value" json:"value"`
	Contract    string `bson:"contract" json:"contract"`
	Method      string `bson:"method" json:"method"`
	// If the token can't be recognized we give it "unknown" method and attach the input data
	Data string `json:"data,omitempty" bson:"data,omitempty"`
}

type RawTxReceipt struct {
	TransactionHash   string  `json:"transactionHash"`
	TransactionIndex  string  `json:"transactionIndex"`
	BlockNumber       string  `json:"blockNumber"`
	BlockHash         string  `json:"blockHash"`
	CumulativeGasUsed string  `json:"cumulativeGasUsed"`
	GasUsed           string  `json:"gasUsed"`
	ContractAddress   string  `json:"contractAddress"`
	Logs              []TxLog `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Status            string  `json:"status"`
}

func (rtr *RawTxReceipt) Convert() TxReceipt {
	return TxReceipt{
		TransactionHash:   rtr.TransactionHash,
		TransactionIndex:  rtr.TransactionIndex,
		BlockNumber:       util.DecodeHex(rtr.BlockNumber),
		BlockHash:         rtr.BlockHash,
		CumulativeGasUsed: util.DecodeHex(rtr.CumulativeGasUsed),
		GasUsed:           util.DecodeHex(rtr.GasUsed),
		ContractAddress:   rtr.ContractAddress,
		Logs:              rtr.Logs,
		LogsBloom:         rtr.LogsBloom,
		Status:            rtr.Status,
	}
}

type TxReceipt struct {
	TransactionHash   string  `json:"transactionHash"`
	TransactionIndex  string  `json:"transactionIndex"`
	BlockNumber       uint64  `json:"blockNumber"`
	BlockHash         string  `json:"blockHash"`
	CumulativeGasUsed uint64  `json:"cumulativeGasUsed"`
	GasUsed           uint64  `json:"gasUsed"`
	ContractAddress   string  `json:"contractAddress"`
	Logs              []TxLog `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Status            string  `json:"status"`
}

// FIXME: this is broken, also probably useless, setting everything to string
type TxLog struct {
	Address          string   `bson:"address" json:"address"`
	Topics           []string `bson:"topics" json:"topics"`
	Data             string   `bson:"data" json:"data"`
	BlockNumber      string   `bson:"blockNumber" json:"blockNumber"`
	TransactionIndex string   `bson:"transactionIndex" json:"transactionIndex"`
	TransactionHash  string   `bson:"transactionHash" json:"transactionHash"`
	BlockHash        string   `bson:"blockHash" json:"blockHash"`
	LogIndex         string   `bson:"logIndex" json:"logIndex"`
	Removed          bool     `bson:"removed" json:"removed"`
}
