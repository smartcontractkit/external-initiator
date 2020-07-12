package blockchain

import (
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

func Test_NEARManager_GetTestJson(t *testing.T) {
	type args struct {
		connectionType subscriber.Type
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"returns JSON when using RPC",
			args{connectionType: subscriber.RPC},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"status"}`),
		},
		{
			"returns empty when using WS",
			args{connectionType: subscriber.WS},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NEARManager{connectionType: tt.args.connectionType}
			if got := m.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NEARManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		connectionType subscriber.Type
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
			fields{connectionType: subscriber.WS},
			args{},
			false,
		},
		{
			"fails unmarshal payload",
			fields{connectionType: subscriber.RPC},
			args{[]byte(`error`)},
			true,
		},
		{
			"fails unmarshal result",
			fields{connectionType: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NEARManager{connectionType: tt.fields.connectionType}
			if err := m.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_NEARManager_GetTriggerJson(t *testing.T) {
	type args struct {
		connectionType subscriber.Type
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			"returns JSON-RPC message when using RPC",
			args{connectionType: subscriber.RPC},
			true,
		},
		{
			"returns empty when using WS",
			args{connectionType: subscriber.WS},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NEARManager{connectionType: tt.args.connectionType}
			if got := m.GetTriggerJson(); (got != nil) != tt.want {
				t.Errorf("GetTriggerJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NEARManager_ParseResponse(t *testing.T) {
}
