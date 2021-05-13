package blockchain

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

func TestCreateKlaytnFilterMessage(t *testing.T) {
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
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"address only",
			store.EthSubscription{Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[null]}]}`),
		},
		{
			"single topic",
			store.EthSubscription{Topics: []string{"abc"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc"]]}]}`),
		},
		{
			"multiple topics",
			store.EthSubscription{Topics: []string{"abc", "def", ""}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
		{
			"address multiple topics",
			store.EthSubscription{Topics: []string{"abc", "def"}, Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
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
			if got := createKlaytnManager(tt.p, store.Subscription{Ethereum: tt.args}).GetTriggerJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}
		})
	}

	t.Run("has invalid filter query", func(t *testing.T) {
		blockHash := common.HexToHash("0xabc")
		got := klaytnManager{ethManager{fq: &filterQuery{BlockHash: &blockHash, FromBlock: "0x1", ToBlock: "0x2"}}}.GetTriggerJson()
		if got != nil {
			t.Errorf("GetTriggerJson() = %s, want nil", got)
		}
	})
}

func TestKlaytnManager_GetTestJson(t *testing.T) {
	type fields struct {
		fq *filterQuery
		p  subscriber.Type
	}
	tests := []struct {
		name   string
		fields fields
		want   []byte
	}{
		{
			"returns JSON when using RPC",
			fields{
				p: subscriber.RPC,
			},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"klay_blockNumber"}`),
		},
		{
			"returns empty when using WS",
			fields{
				p: subscriber.WS,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := klaytnManager{
				ethManager{
					fq: tt.fields.fq,
					p:  tt.fields.p,
				},
			}
			if got := e.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestKlaytnManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		fq *filterQuery
		p  subscriber.Type
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		wantErr           bool
		expectedFromBlock string
	}{
		{
			"does nothing for WS",
			fields{fq: &filterQuery{}, p: subscriber.WS},
			args{},
			false,
			"",
		},
		{
			"parses RPC responses",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`)},
			false,
			"0x1",
		},
		{
			"fails unmarshal payload",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{[]byte(`error`)},
			true,
			"",
		},
		{
			"fails unmarshal result",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := klaytnManager{
				ethManager{
					fq: tt.fields.fq,
					p:  tt.fields.p,
				},
			}
			if err := e.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if e.fq.FromBlock != tt.expectedFromBlock {
				t.Errorf("FromBlock = %s, expected %s", e.fq.FromBlock, tt.expectedFromBlock)
			}
		})
	}
}

func TestKlaytnManager_ParseResponse(t *testing.T) {
	type fields struct {
		fq *filterQuery
		p  subscriber.Type
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name              string
		fields            fields
		args              args
		want              []subscriber.Event
		want1             bool
		expectedFromBlock string
	}{
		{
			"fails parsing invalid payload",
			fields{fq: &filterQuery{}, p: subscriber.WS},
			args{data: []byte(`invalid`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe payload",
			fields{fq: &filterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":[]}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe",
			fields{fq: &filterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":[]}}`)},
			nil,
			false,
			"",
		},
		{
			"successfully parses WS response",
			fields{fq: &filterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":{"data":"test"}}}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","blockNumber":"","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"",
		},
		{
			"fails parsing invalid RPC payload",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid block number in RPC event payload",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","blockNumber":"","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"",
		},
		{
			"updates fromBlock from RPC payload",
			fields{fq: &filterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test","blockNumber":"0x0"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","blockNumber":"0x0","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"0x1",
		},
		{
			"does not update fromBlock in the past from RPC payload",
			fields{fq: &filterQuery{FromBlock: "0x1"}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test","blockNumber":"0x0"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","blockNumber":"0x0","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"0x1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := klaytnManager{
				ethManager{
					fq: tt.fields.fq,
					p:  tt.fields.p,
				},
			}
			got, got1 := e.ParseResponse(tt.args.data)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseResponse() got = %s, want %v", got, tt.want)
			}
			if got1 != tt.want1 {
				t.Errorf("ParseResponse() got1 = %v, want %v", got1, tt.want1)
			}
			if e.fq.FromBlock != tt.expectedFromBlock {
				t.Errorf("FromBlock = %s, expected %s", e.fq.FromBlock, tt.expectedFromBlock)
			}
		})
	}
}
