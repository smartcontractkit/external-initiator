package blockchain

import (
	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"reflect"
	"testing"
)

func TestCreateEthFilterMessage(t *testing.T) {
	type args struct {
		addressesStr []string
		topicsStr    []string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"empty",
			args{},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"address only",
			args{addressesStr: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"single topic",
			args{topicsStr: []string{"abc"}},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc"]]}]}`),
		},
		{
			"multiple topics",
			args{topicsStr: []string{"abc", "def"}},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
		{
			"address multiple topics",
			args{addressesStr: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}, topicsStr: []string{"abc", "def"}},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"eth_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CreateEthFilterMessage(tt.args.addressesStr, tt.args.topicsStr).Json(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CreateEthFilterMessage() = %s, want %s", got, tt.want)
			}
		})
	}
}

func Test_toBlockNumArg(t *testing.T) {
	type args struct {
		number *big.Int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"empty should give latest", args{nil}, "latest"},
		{"int should be converted to hex", args{big.NewInt(3234512922)}, "0xc0cac01a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toBlockNumArg(tt.args.number); got != tt.want {
				t.Errorf("toBlockNumArg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_toFilterArg(t *testing.T) {
	type args struct {
		q ethereum.FilterQuery
	}

	blockHash := common.HexToHash("abc")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"cannot specify both Blockhash and FromBlock",
			args{ethereum.FilterQuery{
				BlockHash: &blockHash,
				FromBlock: big.NewInt(3234512922),
			}},
			true,
		},
		{
			"cannot specify both Blockhash and ToBlock",
			args{ethereum.FilterQuery{
				BlockHash: &blockHash,
				ToBlock:   big.NewInt(3234512922),
			}},
			true,
		},
		{
			"regular query passes",
			args{ethereum.FilterQuery{
				Addresses: []common.Address{},
				Topics:    [][]common.Hash{},
			}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toFilterArg(tt.args.q)
			if (err != nil) != tt.wantErr {
				t.Errorf("toFilterArg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
