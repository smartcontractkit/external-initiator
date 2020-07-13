package blockchain

import (
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

func Test_nearManager_GetTestJson(t *testing.T) {
	type args struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"returns JSON when using RPC",
			args{filter: filter, connectionType: subscriber.RPC},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"status"}`),
		},
		{
			"returns empty when using WS",
			args{filter: filter, connectionType: subscriber.WS},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: filter, connectionType: tt.args.connectionType}
			if got := m.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nearManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	type args struct {
		data []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"does nothing for WS",
			fields{filter: filter, connectionType: subscriber.WS},
			args{},
			false,
		},
		{
			"fails unmarshal payload",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`error`)},
			true,
		},
		{
			"fails unmarshal result",
			fields{filter: filter, connectionType: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: filter, connectionType: tt.fields.connectionType}
			if err := m.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_nearManager_GetTriggerJson(t *testing.T) {
	type args struct {
		filter         nearFilter
		connectionType subscriber.Type
	}
	filter := nearFilter{
		JobID:      "job#1",
		AccountIDs: []string{"oracle.chainlink.testnet"},
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"returns JSON-RPC message when using RPC",
			args{filter: filter, connectionType: subscriber.RPC},
			true,
		},
		{
			"returns empty when using WS",
			args{filter: filter, connectionType: subscriber.WS},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := nearManager{filter: tt.args.filter, connectionType: tt.args.connectionType}
			if got := m.GetTriggerJson(); (got != nil) != tt.want {
				t.Errorf("GetTriggerJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_nearManager_ParseResponse(t *testing.T) {
}
