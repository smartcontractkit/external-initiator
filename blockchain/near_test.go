package blockchain

import (
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

func Test_NEARManager_GetTestJson(t *testing.T) {
	type args struct {
		p subscriber.Type
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			"returns JSON when using RPC",
			args{
				p: subscriber.RPC,
			},
			[]byte(`{"jsonrpc":"2.0","id":1,"method":"status"}`),
		},
		{
			"returns empty when using WS",
			args{
				p: subscriber.WS,
			},
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NEARManager{
				p: tt.args.p,
			}
			if got := m.GetTestJson(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetTestJson() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_NEARManager_ParseTestResponse(t *testing.T) {
	type fields struct {
		p subscriber.Type
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
			fields{p: subscriber.WS},
			args{},
			false,
		},
		{
			"fails unmarshal payload",
			fields{p: subscriber.RPC},
			args{[]byte(`error`)},
			true,
		},
		{
			"fails unmarshal result",
			fields{p: subscriber.RPC},
			args{[]byte(`{"jsonrpc":"2.0","id":1,"result":["0x1"]}`)},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NEARManager{
				p: tt.fields.p,
			}
			if err := m.ParseTestResponse(tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("ParseTestResponse() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
