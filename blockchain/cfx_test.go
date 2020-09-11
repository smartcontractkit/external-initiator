package blockchain

import (
	"reflect"
	"testing"

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
			"empty RPC",
			store.CfxSubscription{},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_getLogs","params":[{"address":null,"fromEpoch":"latest_state","toEpoch":"latest_state","topics":[null]}]}`),
		},
		{
			"RPC address multiple topics",
			store.CfxSubscription{Topics: []string{"abc", "def"}, Addresses: []string{"0x049Bd8C3adC3fE7d3Fc2a44541d955A537c2A484"}},
			subscriber.RPC,
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"cfx_getLogs","params":[{"address":["0x049bd8c3adc3fe7d3fc2a44541d955a537c2a484"],"fromEpoch":"latest_state","toEpoch":"latest_state","topics":[["0x0000000000000000000000000000000000000000000000000000000000000abc","0x0000000000000000000000000000000000000000000000000000000000000def"]]}]}`),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := createCfxManager(tt.p, store.Subscription{Conflux: tt.args}).GetTriggerJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTriggerJson() = %s, want %s", got, tt.want)
			}
		})
	}

	t.Run("has invalid filter query", func(t *testing.T) {
		blockHash := common.HexToHash("0xabc")
		got := cfxManager{fq: &cfxFilterQuery{BlockHash: &blockHash, fromEpoch: "0x1", toEpoch: "0x2"}}.GetTriggerJson()
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
			if e.fq.fromEpoch != tt.expectedFromEpoch {
				t.Errorf("FromEpoch = %s, expected %s", e.fq.fromEpoch, tt.expectedFromEpoch)
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
			"fails parsing invalid RPC payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":{}}`)},
			nil,
			false,
			"",
		},
		{
			"fails parsing invalid block number in RPC event payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","epochNumber":"","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"",
		},
		{
			"updates fromEpoch from RPC payload",
			fields{fq: &cfxFilterQuery{}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test","epochNumber":"0x0"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","epochNumber":"0x0","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
			true,
			"0x1",
		},
		{
			"does not update fromEpoch in the past from RPC payload",
			fields{fq: &cfxFilterQuery{fromEpoch: "0x1"}, p: subscriber.RPC},
			args{data: []byte(`{"jsonrpc":"2.0","id":1,"result":[{"data":"test","epochNumber":"0x0"}]}`)},
			[]subscriber.Event{subscriber.Event(`{"logIndex":"","epochNumber":"0x0","blockHash":"","transactionHash":"","transactionIndex":"","address":"","data":"test","topics":null}`)},
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
			if e.fq.fromEpoch != tt.expectedFromEpoch {
				t.Errorf("fromEpoch = %s, expected %s", e.fq.fromEpoch, tt.expectedFromEpoch)
			}
		})
	}
}

func Test_cfxFilterQuery_toMapInterface(t *testing.T) {
	type fields struct {
		BlockHash *common.Hash
		fromEpoch string
		toEpoch   string
		Addresses []common.Address
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
				"address":   []common.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty toEpoch",
			fields{toEpoch: "0x1"},
			map[string]interface{}{
				"address":   []common.Address{},
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
				"address":   []common.Address{},
				"topics":    [][]common.Hash{},
				"fromEpoch": "0x0",
				"toEpoch":   "latest_state",
			},
			false,
		},
		{
			"uses non-empty fromEpoch",
			fields{fromEpoch: "0x1"},
			map[string]interface{}{
				"address":   []common.Address{},
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
				fromEpoch: tt.fields.fromEpoch,
				toEpoch:   tt.fields.toEpoch,
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
