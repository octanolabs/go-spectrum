package rpc

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"

	"net/http"
	"time"

	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/util"
)

// TODO: this should use go-ubiq/rpc

type Config struct {
	Url     string
	Timeout string
}

type RPCClient struct {
	Url    string
	client *http.Client
}

type JSONRpcResp struct {
	Id     *json.RawMessage       `json:"id"`
	Result *json.RawMessage       `json:"result"`
	Error  map[string]interface{} `json:"error"`
}

func NewRPCClient(cfg *Config) *RPCClient {
	rpcClient := &RPCClient{Url: cfg.Url}

	timeoutIntv, err := time.ParseDuration(cfg.Timeout)
	if err != nil {
		log.Fatalf("RPC: can't parse duration: %v", err)
	}

	rpcClient.client = &http.Client{
		Timeout: timeoutIntv,
	}
	return rpcClient
}

func (r *RPCClient) doPost(method string, params interface{}) (*JSONRpcResp, error) {
	jq := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  method,
		"params":  params,
		"id":      0,
	}

	data, err := json.Marshal(jq)

	if err != nil {
		log.Debugf("Error marshalling json (doPost): %v", err)
		return nil, err
	}

	req, err := http.NewRequest("POST", r.Url, bytes.NewBuffer(data))

	if err != nil {
		log.Debugf("Error creating http req: %v", err)
		return nil, err
	}

	req.Header.Set("Content-Length", (string)(len(data)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var rpcResp *JSONRpcResp
	err = json.NewDecoder(resp.Body).Decode(&rpcResp)
	if err != nil {
		return nil, err
	}
	if rpcResp.Error != nil {
		return nil, errors.New(rpcResp.Error["message"].(string))
	}
	return rpcResp, err
}

func (r *RPCClient) getBlockBy(method string, params []interface{}) (*models.Block, error) {
	rpcResp, err := r.doPost(method, params)
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *models.RawBlock
		err = json.Unmarshal(*rpcResp.Result, &reply)

		return reply.Convert(), err
	}
	return nil, nil
}

func (r *RPCClient) GetLatestBlock() (*models.Block, error) {
	bn, err := r.LatestBlockNumber()

	if err != nil {
		return nil, err
	}

	params := []interface{}{fmt.Sprintf("0x%x", bn)}
	return r.getBlockBy("eth_getBlockByNumber", params)
}

func (r *RPCClient) GetBlockByHeight(height uint64) (*models.Block, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), true}
	return r.getBlockBy("eth_getBlockByNumber", params)
}

func (r *RPCClient) GetBlockByHash(hash string) (*models.Block, error) {
	params := []interface{}{hash, true}
	return r.getBlockBy("eth_getBlockByHash", params)
}

func (r *RPCClient) getUncleBy(method string, params []interface{}) (*models.Uncle, error) {
	rpcResp, err := r.doPost(method, params)
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *models.RawUncle
		err = json.Unmarshal(*rpcResp.Result, &reply)

		return reply.Convert(), err
	}
	return nil, nil
}

func (r *RPCClient) GetUncleByBlockNumberAndIndex(height uint64, index int) (*models.Uncle, error) {
	params := []interface{}{fmt.Sprintf("0x%x", height), fmt.Sprintf("0x%x", index)}
	return r.getUncleBy("eth_getUncleByBlockNumberAndIndex", params)
}

func (r *RPCClient) LatestBlockNumber() (uint64, error) {
	rpcResp, err := r.doPost("eth_blockNumber", []interface{}{})

	if err != nil {
		return 0, err
	}

	if rpcResp.Result != nil {
		var reply string
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return util.DecodeHex(reply), err
	}
	return 0, nil

}

func (r *RPCClient) GetTxReceipt(hash string) (*models.TxReceipt, error) {
	rpcResp, err := r.doPost("eth_getTransactionReceipt", []string{hash})
	if err != nil {
		return nil, err
	}
	if rpcResp.Result != nil {
		var reply *models.RawTxReceipt
		err = json.Unmarshal(*rpcResp.Result, &reply)
		return reply.Convert(), err
	}
	return nil, nil
}

func (r *RPCClient) Ping() error {
	_, err := r.doPost("web3_clientVersion", []string{})
	if err != nil {
		return err
	}
	return nil
}
