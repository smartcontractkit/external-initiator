package blockchain

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"math/big"
	"reflect"
	"testing"
)

func TestCreateEthFilterMessage(t *testing.T) {
	tests := []struct {
		name string
		args store.EthSubscription
		p    subscriber.Type
		want []byte
	}{
		{
			"empty",
			store.EthSubscription{},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"address only",
			store.EthSubscription{Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"single topic",
			store.EthSubscription{Topics: []string{"abc"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc"]]}]}`),
		},
		{
			"multiple topics",
			store.EthSubscription{Topics: []string{"abc", "def"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
		{
			"address multiple topics",
			store.EthSubscription{Topics: []string{"abc", "def"}, Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
		{
			"empty RPC",
			store.EthSubscription{},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"address":null,"fromBlock":"latest","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"RPC address multiple topics",
			store.EthSubscription{Topics: []string{"abc", "def"}, Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_getLogs","params":[{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"latest","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateEthManager(tt.p, tt.args).GetTriggerJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_toFilterArg(t *testing.T) {
	type args struct {
		q filterQuery
	}

	blockHash := common.HexToHash("abc")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cannot specify both Blockhash and FromBlock",
			args{filterQuery{
				BlockHash: &blockHash,
				FromBlock: hexutil.EncodeBig(big.NewInt(3234512922)),
			}},
			true,
		},
		{
			"cannot specify both Blockhash and ToBlock",
			args{filterQuery{
				BlockHash: &blockHash,
				ToBlock:   hexutil.EncodeBig(big.NewInt(3234512922)),
			}},
			true,
		},
		{
			"regular query passes",
			args{filterQuery{
				Addresses: []common.Address{},
				Topics:    [][]common.Hash{},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := tt.args.q.toMapInterface()
			if (err != nil) != tt.wantErr {
				t.Errorf("toFilterArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
