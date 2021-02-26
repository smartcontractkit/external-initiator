package blockchain

import (
	"reflect"
	"testing"

	"github.com/Conflux-Chain/go-conflux-sdk/types/cfxaddress"

	"github.com/ethereum/go-ethereum/common"
	"github.com/magiconair/properties/assert"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

func TestCreateCfxFilterMessage(t *testing.T) {
	tests := []struct {
		name string
		args store.CfxSubscription
		p    subscriber.Type
		want []byte
	}{
		{
			"empty",
			store.CfxSubscription{},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_subscribe","params":["logs",{"address":null,"fromEpoch":"0x0","toEpoch":"latest_state","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
		{
			"address only",
			store.CfxSubscription{Addresses: []string{"cfxtest:acdjv47k166p1pt4e8yph9rbcumrpbn2u69wyemxv0"}},
			subscriber.WS,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_subscribe","params":["logs",{"address":["CFXTEST:TYPE.CONTRACT:ACDJV47K166P1PT4E8YPH9RBCUMRPBN2U69WYEMXV0"],"fromEpoch":"0x0","toEpoch":"latest_state","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
		{
			"empty RPC",
			store.CfxSubscription{},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_getLogs","params":[{"address":null,"fromEpoch":"latest_state","toEpoch":"latest_state","topics":[["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65"],["0x0000000000000000000000000000000000000000000000000000000000000000"]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createCfxManager(tt.p, store.Subscription{Conflux: tt.args}).GetTriggerJson()
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}
		})
	}

	t.Run("has invalid filter query", func(t *testing.T) {
		blockHash := common.HexToHash("0xabc")
		got := cfxManager{fq: &cfxFilterQuery{BlockHash: &blockHash, FromEpoch: "0x1", ToEpoch: "0x2"}}.GetTriggerJson()
		if got != nil {
			t.Errorf("GetTriggerJson() = %s, want nil", got)
		}
	})
}

func TestCfxManager_GetTestJson(t *testing.T) {
	type fields struct {
		fq *cfxFilterQuery
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
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_epochNumber"}`),
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
			e := cfxManager{
				fq: tt.fields.fq,
				p:  tt.fields.p,
			}
			if got := e.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCfxManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		fq *cfxFilterQuery
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
		expectedFromEpoch string
	}{
		{
			"does nothing for WS",
			fields{fq: &cfxFilterQuery{}, p: subscriber.WS},
			args{},
			false,
			"",
		},
		{
			"parses RPC responses",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":"0x1"}`)},
			false,
			"0x1",
		},
		{
			"fails unmarshal payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{[]byte(`error`)},
			true,
			"",
		},
		{
			"fails unmarshal result",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := cfxManager{
				fq: tt.fields.fq,
				p:  tt.fields.p,
			}
			if err := e.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
			if e.fq.FromEpoch != tt.expectedFromEpoch {
				t.Errorf("FromEpoch = %s, expected %s", e.fq.FromEpoch, tt.expectedFromEpoch)
			}
		})
	}
}

func TestCfxManager_ParseResponse(t *testing.T) {
	type fields struct {
		fq *cfxFilterQuery
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
		expectedFromEpoch string
	}{
		{
			"fails parsing invalid payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.WS},
			args{data: []byte(`invalid`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":[]}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid WS subscribe",
			fields{fq: &cfxFilterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":[]}}`)},
			nil,
			false,
			"",
		},
		{
			"successfully parses WS Oracle request",
			fields{fq: &cfxFilterQuery{}, p: subscriber.WS},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"params":{"subscription":"test","result":{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"CFXTEST:TYPE.CONTRACT:ACFR9765YBHVRE6GPVZEHBY5P43329UJNAN8GFR20F","logIndex":"0x0","epochNumber":"0x2","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}}}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0x8adFf79Ba04F169386646A43869b66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"",
		},
		{
			"fails parsing invalid RPC payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid epoch number in RPC event payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"0xFadfF79bA04F169386646a43869B66B39c7E0858","logIndex":"0x0","epochNumber":"abc","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			nil,
			false,
			"",
		},
		{
			"updates fromEpoch from RPC payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"CFXTEST:TYPE.CONTRACT:ACFR9765YBHVRE6GPVZEHBY5P43329UJNAN8GFR20F","logIndex":"0x0","epochNumber":"0x3","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0x8adFf79Ba04F169386646A43869b66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"0x4",
		},
		{
			"does not update fromEpoch in the past from RPC payload",
			fields{fq: &cfxFilterQuery{FromEpoch: "0x1"}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"0x0000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b00000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000100000000000000000000000000000000000000000000000000000000000000005663676574783f68747470733a2f2f6d696e2d6170692e63727970746f636f6d706172652e636f6d2f646174612f70726963653f6673796d3d455448267473796d733d5553446470617468635553446574696d65731864","address":"CFXTEST:TYPE.CONTRACT:ACFR9765YBHVRE6GPVZEHBY5P43329UJNAN8GFR20F","logIndex":"0x0","epochNumber":"0x0","blockHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionHash":"0xabc0000000000000000000000000000000000000000000000000000000000000","transactionIndex":"0x0","topics":["0xd8d7ecc4800d25fa53ce0372f13a416d98907a7ef3d8d3bdd79cf4fe75529c65","0x0000000000000000000000000000000000000000000000000000000000000000"]}]}`)},
			[]subscriber.Event{subscriber.Event(`{"address":"0x8adFf79Ba04F169386646A43869b66B39c7E0858","dataPrefix":"0x354f99e2ac319d0d1ff8975c41c72bf347fb69a4874e2641bd19c32e09eb88b80000000000000000000000000000000000000000000000000de0b6b3a76400000000000000000000000000007d0965224facd7156df0c9a1adf3a94118026eeb92cdaaf300000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000005ef1cd6b","functionSelector":"0x4ab0d190","get":"https://min-api.cryptocompare.com/data/price?fsym=ETH\u0026tsyms=USD","path":"USD","times":100}`)},
			true,
			"0x1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := cfxManager{
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
			if e.fq.FromEpoch != tt.expectedFromEpoch {
				t.Errorf("FromEpoch = %s, expected %s", e.fq.FromEpoch, tt.expectedFromEpoch)
			}
		})
	}
}

func Test_cfxFilterQuery_toMapInterface(t *testing.T) {
	type fields struct {
		BlockHash *common.Hash
		FromEpoch string
		ToEpoch   string
		Addresses []cfxaddress.Address
		Topics    [][]common.Hash
	}
	tests := []struct {
		name    string
		fields  fields
		want    map[string]interface{}
		wantErr bool
	}{
		{
			"empty toEpoch becomes latest_state",
			fields{},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty toEpoch",
			fields{ToEpoch: "0x1"},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "0x1",
			},
			false,
		},
		{
			"empty fromEpoch becomes 0x0",
			fields{},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty fromEpoch",
			fields{FromEpoch: "0x1"},
			map[string]interface{}{
				"address":   []cfxaddress.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x1",
				"toEpoch":   "latest_state",
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := cfxFilterQuery{
				BlockHash: tt.fields.BlockHash,
				FromEpoch: tt.fields.FromEpoch,
				ToEpoch:   tt.fields.ToEpoch,
				Addresses: tt.fields.Addresses,
				Topics:    tt.fields.Topics,
			}
			got, err := q.toMapInterface()
			if (err != nil) != tt.wantErr {
				t.Errorf("toMapInterface() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			mapInterface, ok := got.(map[string]interface{})
			require.True(t, ok)
			keys := []string{"fromEpoch", "toEpoch"}
			for _, key := range keys {
				assert.Equal(t, mapInterface[key], tt.want[key])
			}
		})
	}
}
