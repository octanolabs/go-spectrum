package rpc

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/ubiq/go-ubiq/v6/log"

	"github.com/mitchellh/go-homedir"
	"github.com/ubiq/go-ubiq/v6/common/hexutil"
	"github.com/ubiq/go-ubiq/v6/rpc"

	"github.com/octanolabs/go-spectrum/models"
	"github.com/octanolabs/go-spectrum/util"
)

type Config struct {
	Type     string `json:"type"`
	Endpoint string `json:"endpoint"`
}

type RPCClient struct {
	client *rpc.Client
}

func dialNewClient(cfg *Config) (*rpc.Client, error) {

	var (
		client *rpc.Client
		err    error
	)

	switch cfg.Type {
	case "http":
		if client, err = rpc.DialHTTP(cfg.Endpoint); err != nil {
			return nil, err
		}
	case "unix", "ipc":
		if client, err = rpc.DialIPC(context.Background(), cfg.Endpoint); err != nil {
			return nil, err
		}
	case "ws", "websocket", "websockets":
		if client, err = rpc.DialWebsocket(context.Background(), cfg.Endpoint, ""); err != nil {
			return nil, err
		}
	default:
		fp, err := homedir.Expand("~/.ubiq/gubiq.ipc")
		if err != nil {
			return nil, err
		}
		if client, err = rpc.DialIPC(context.Background(), fp); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func NewRPCClient(cfg *Config) *RPCClient {

	client, err := dialNewClient(cfg)
	if err != nil {
		log.Error("could not dial rpc client", "err", err)
		os.Exit(1)
	}

	rpcClient := &RPCClient{client}

	return rpcClient
}

func (r *RPCClient) getBlockBy(method string, params ...interface{}) (models.Block, error) {
	var reply models.RawBlock

	err := r.client.Call(&reply, method, params...)

	if err != nil {
		return models.Block{}, err
	}

	return reply.Convert(), nil
}

func (r *RPCClient) getUncleBy(method string, params ...interface{}) (models.Uncle, error) {
	var reply models.RawUncle

	err := r.client.Call(&reply, method, params...)
	if err != nil {
		return models.Uncle{}, err
	}

	return reply.Convert(), nil
}

func (r *RPCClient) GetLatestBlock() (models.Block, error) {
	bn, err := r.LatestBlockNumber()

	if err != nil {
		return models.Block{}, err
	}

	return r.getBlockBy("eth_getBlockByNumber", hexutil.EncodeUint64(bn))
}

func (r *RPCClient) GetBlockByHeight(height uint64) (models.Block, error) {
	return r.getBlockBy("eth_getBlockByNumber", hexutil.EncodeUint64(height), true)
}

func (r *RPCClient) GetBlockByHash(hash string) (models.Block, error) {
	return r.getBlockBy("eth_getBlockByHash", hash, true)
}

func (r *RPCClient) GetUncleByBlockNumberAndIndex(height uint64, index int) (models.Uncle, error) {
	return r.getUncleBy("eth_getUncleByBlockNumberAndIndex", hexutil.EncodeUint64(height), hexutil.EncodeUint64(uint64(index)))
}

// What is the purpose of this method

func (r *RPCClient) GetUnclesInBlock(uncles []string, height uint64) ([]models.Uncle, error) {

	var u []models.Uncle

	for k := range uncles {
		uncle, err := r.GetUncleByBlockNumberAndIndex(height, k)
		if err != nil {
			return u, errors.New("Error getting uncle: " + err.Error())
		}
		u = append(u, uncle)
	}
	return u, nil
}

func (r *RPCClient) LatestBlockNumber() (uint64, error) {
	var bn string

	err := r.client.Call(&bn, "eth_blockNumber")
	if err != nil {
		return 0, err
	}

	return util.DecodeHex(bn), nil

}

func (r *RPCClient) GetTxReceipt(hash string) (models.TxReceipt, error) {
	var reply models.RawTxReceipt

	err := r.client.Call(&reply, "eth_getTransactionReceipt", hash)

	if err != nil {
		return models.TxReceipt{}, err
	}

	return reply.Convert(), nil
}

func (r *RPCClient) Ping() (string, error) {
	var version string

	err := r.client.Call(&version, "web3_clientVersion")
	if err != nil {
		return "", err
	}

	return version, nil
}

func (r *RPCClient) TraceTransaction(hash string) (models.InternalTx, error) {
	var trace models.RawTxTrace

	c, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	defer cancel()

	err := r.client.CallContext(c, &trace, "debug_traceTransaction", hash, map[string]interface{}{
		"tracer": "callTracer",
		//if tracer errors out with "execution timeout" increase this timeout
		"timeout": "300s",
	})
	if err != nil {
		return models.InternalTx{}, err
	}

	return trace.Convert(), nil
}

func (r *RPCClient) GetState(blockNumber uint64) (models.RawState, error) {
	var state models.RawState

	err := r.client.Call(&state, "debug_dumpBlock", hexutil.EncodeUint64(blockNumber))

	if err != nil {
		return models.RawState{}, err
	}

	return state, nil
}
