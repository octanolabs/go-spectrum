package models

import (
	"github.com/octanolabs/go-spectrum/util"
	"github.com/ubiq/go-ubiq/v6/log"
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
	Status           bool   `json:"status"`
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

// generic function call on contract that is not a token transfer
func (tx *Transaction) IsContractCall() bool {
	return !tx.IsTokenTransfer() && tx.ContractAddress == "" && tx.Input != "0x"
}

func (tx *Transaction) IsContractDeployTxn() bool {
	return tx.ContractAddress != ""
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
		log.Error("couldn't proces token transfer: input length is not standard", "len", len(tx.Input))
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
	Status      bool   `json:"status"`
	// If the token can't be recognized we give it "unknown" method and attach the input data
	Data string `json:"data,omitempty" bson:"data,omitempty"`
}

type RawTxReceipt struct {
	BlockHash         string  `json:"blockHash"`
	BlockNumber       string  `json:"blockNumber"`
	ContractAddress   string  `json:"contractAddress"`
	CumulativeGasUsed string  `json:"cumulativeGasUsed"`
	From              string  `json:"from"`
	GasUsed           string  `json:"gasUsed"`
	Logs              []TxLog `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Status            string  `json:"status"`
	To                string  `json:"to"`
	TransactionHash   string  `json:"transactionHash"`
	TransactionIndex  string  `json:"transactionIndex"`
}

//{
//"blockHash": "0x5cae9e3aae39e3e9970de0307445fd442333713f9e6f5d0fed51e19b56ff6259",
//"blockNumber": "0x117640",
//"contractAddress": null,
//"cumulativeGasUsed": "0x47c70",
//"from": "0x09692a71d42c209f42b731af8edd6910287437d3",
//"gasUsed": "0x5208",
//"logs": [],
//"logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
//"status": "0x1",
//"to": "0xb3c4e9ca7c12a6277deb9eef2dece65953d1c864",
//"transactionHash": "0x8d028a79d671bef7e0b405eb9905978186421611d895103c00244a89dc3a13c3",
//"transactionIndex": "0xd"
//}

func (rtr *RawTxReceipt) Convert() TxReceipt {
	var status bool

	if rtr.Status == "0x1" {
		status = true
	}

	return TxReceipt{
		BlockNumber:       util.DecodeHex(rtr.BlockNumber),
		BlockHash:         rtr.BlockHash,
		ContractAddress:   rtr.ContractAddress,
		CumulativeGasUsed: util.DecodeHex(rtr.CumulativeGasUsed),
		From:              rtr.From,
		GasUsed:           util.DecodeHex(rtr.GasUsed),
		Logs:              rtr.Logs,
		LogsBloom:         rtr.LogsBloom,
		Status:            status,
		To:                rtr.To,
		TransactionHash:   rtr.TransactionHash,
		TransactionIndex:  rtr.TransactionIndex,
	}
}

type TxReceipt struct {
	BlockHash         string  `json:"blockHash"`
	BlockNumber       uint64  `json:"blockNumber"`
	ContractAddress   string  `json:"contractAddress"`
	CumulativeGasUsed uint64  `json:"cumulativeGasUsed"`
	From              string  `json:"from"`
	GasUsed           uint64  `json:"gasUsed"`
	Logs              []TxLog `json:"logs"`
	LogsBloom         string  `json:"logsBloom"`
	Status            bool    `json:"status"`
	To                string  `json:"to"`
	TransactionHash   string  `json:"transactionHash"`
	TransactionIndex  string  `json:"transactionIndex"`
}

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

// RawTxTrace is what we get from the node tracer
type RawTxTrace struct {
	Type    string       `json:"type"`
	From    string       `json:"from"`
	To      string       `json:"to"`
	Value   string       `json:"value,omitempty"` // hex
	Gas     string       `json:"gas"`             // hex
	GasUsed string       `json:"gasUsed"`         // hex
	Input   string       `json:"input"`
	Output  string       `json:"output"`
	Time    string       `json:"-"`
	Calls   []RawTxTrace `json:"calls,omitempty"`
}

func (rtt *RawTxTrace) Convert() InternalTx {
	t := InternalTx{
		Type:    rtt.Type,
		From:    rtt.From,
		To:      rtt.To,
		Value:   util.DecodeValueHex(rtt.Value),
		Gas:     util.DecodeValueHex(rtt.Gas),
		GasUsed: util.DecodeValueHex(rtt.GasUsed),
		Input:   rtt.Input,
		Output:  rtt.Output,
		Calls:   nil,
	}

	if rtt.Calls != nil {
		t.Calls = []InternalTx{}
		for _, nestedTrace := range rtt.Calls {
			t.Calls = append(t.Calls, nestedTrace.Convert())
		}
	}

	return t
}

//TxTrace is a 'meta' schema that includes a trace, the originating txhash and the block number the it was included in
type TxTrace struct {
	OriginTxHash  string `json:"hash" bson:"hash"`
	OriginBlockNo int64  `json:"number" bson:"number"`
	Trace         InternalTx
}

// InternalTx the schema representing the state transitions
type InternalTx struct {
	Type    string       `json:"type" bson:"type"`
	From    string       `json:"from" bson:"from"`
	To      string       `json:"to" bson:"to"`
	Value   string       `json:"value,omitempty" bson:"value,omitempty"` //convert to number -> string
	Gas     string       `json:"gas" bson:"gas"`                         //convert to number -> string
	GasUsed string       `json:"gasUsed" bson:"gasUsed"`                 //convert to number -> string
	Input   string       `json:"input" bson:"input"`
	Output  string       `json:"output" bson:"output"`
	Calls   []InternalTx `json:"calls,omitempty" bson:"calls,omitempty"`
}
