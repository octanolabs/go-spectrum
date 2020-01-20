package rpc

import (
	"context"

	"github.com/mitchellh/go-homedir"
	"github.com/ubiq/go-ubiq/common/hexutil"
	"github.com/ubiq/go-ubiq/rpc"

	log "github.com/sirupsen/logrus"

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
		log.Errorf("Error: could not dial rpc client:%v", err)
	}

	rpcClient := &RPCClient{client: client}

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

func (r *RPCClient) GetUnclesInBlock(uncles []string, height uint64) []models.Uncle {

	var u []models.Uncle

	for k := range uncles {
		uncle, err := r.GetUncleByBlockNumberAndIndex(height, k)
		if err != nil {
			log.Errorf("Error getting uncle: %v", err)
			return u
		}
		u = append(u, uncle)
	}
	return u
}

func (r *RPCClient) LatestBlockNumber() (uint64, error) {
	var bn string

	err := r.client.Call(&bn, "eth_blockNumber", []interface{}{})
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

	return models.TxReceipt{}, nil
}

func (r *RPCClient) Ping() error {
	err := r.client.Call(nil, "web3_clientVersion", []interface{}{})
	if err != nil {
		return err
	}

	return nil
}
