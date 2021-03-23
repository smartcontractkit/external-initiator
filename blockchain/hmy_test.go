package blockchain

import (
	"reflect"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

func TestCreateHmyFilterMessage(t *testing.T) {
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
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"hmy_subscribe","params":["logs",{"address":null,"fromBlock":"0x0","toBlock":"latest","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
		{
			"address only",
			store.EthSubscription{Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"hmy_subscribe","params":["logs",{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromBlock":"0x0","toBlock":"latest","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
		{
			"empty RPC",
			store.EthSubscription{},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"hmy_getLogs","params":[{"address":null,"fromBlock":"latest","toBlock":"latest","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createHmyManager(tt.p, store.Subscription{Ethereum: tt.args}).GetTriggerJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}
		})
	}

	t.Run("has invalid filter query", func(t *testing.T) {
		blockHash := common.HexToHash("0xabc")
		got := hmyManager{fq: &filterQuery{BlockHash: &blockHash, FromBlock: "0x1", ToBlock: "0x2"}}.GetTriggerJson()
		if got != nil {
			t.Errorf("GetTriggerJson() = %s, want nil", got)
		}
	})
}

func TestHmyManager_GetTestJson(t *testing.T) {
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
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"hmy_blockNumber"}`),
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
			e := hmyManager{
				fq: tt.fields.fq,
				p:  tt.fields.p,
			}
			if got := e.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHmyManager_ParseTestResponse(t *testing.T) {
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
			e := hmyManager{
				fq: tt.fields.fq,
				p:  tt.fields.p,
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

func TestHmyManager_ParseResponse(t *testing.T) {
	type manager struct {
		fq *filterQuery
		p  subscriber.Type
	}
	type fields struct {
		manager
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
			fields{manager{fq: &filterQuery{}, p: subscriber.WS}},
			args{data: []byte(`invalid`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe payload",
			fields{manager{fq: &filterQuery{}, p: subscriber.WS}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":[]}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe",
			fields{manager{fq: &filterQuery{}, p: subscriber.WS}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":[]}}`)},
			nil,
			false,
			"",
		},
		{
			"successfully parses WS Oracle request",
			fields{manager{fq: &filterQuery{}, p: subscriber.WS}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","logIndex":"0x0","blockNumber":"0x2","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}}}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"",
		},
		{
			"fails parsing invalid RPC payload",
			fields{manager{fq: &filterQuery{}, p: subscriber.RPC}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid block number in RPC event payload",
			fields{manager{fq: &filterQuery{}, p: subscriber.RPC}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","logIndex":"0x0","blockNumber":"abc","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			nil,
			false,
			"",
		},
		{
			"updates fromBlock from RPC payload",
			fields{manager{fq: &filterQuery{}, p: subscriber.RPC}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","logIndex":"0x0","blockNumber":"0x3","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"0x4",
		},
		{
			"does not update fromBlock in the past from RPC payload",
			fields{manager{fq: &filterQuery{FromBlock: "0x1"}, p: subscriber.RPC}},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","logIndex":"0x0","blockNumber":"0x0","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"0x1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := hmyManager{
				fq: tt.fields.fq,
				p:  tt.fields.p,
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
