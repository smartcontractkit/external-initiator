package blockchain

import (
	"testing"

	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

func Test_GetConnectionType(t *testing.T) {
	type args struct {
		rawUrl string
	}
	tests := []struct {
		name    string
		args    args
		want    subscriber.Type
		wantErr bool
	}{
		{
			"fails on invalid type",
			args{rawUrl: "invalid://localhost/"},
			subscriber.Unknown,
			true,
		},
		{
			"fails on invalid URL",
			args{"http://a b.com/"},
			subscriber.Unknown,
			true,
		},
		{
			"returns WS on ws://",
			args{"ws://localhost/"},
			subscriber.WS,
			false,
		},
		{
			"returns WS on secure wss://",
			args{"wss://localhost/"},
			subscriber.WS,
			false,
		},
		{
			"returns RPC on http://",
			args{"http://localhost/"},
			subscriber.RPC,
			false,
		},
		{
			"returns RPC on secure https://",
			args{"https://localhost/"},
			subscriber.RPC,
			false,
		},
		{
			"returns error on unknown protocol",
			args{"postgres://localhost/"},
			subscriber.Unknown,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetConnectionType(store.Endpoint{Url: tt.args.rawUrl})
			if (err != nil) != tt.wantErr {
				t.Errorf("getConnectionType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("getConnectionType() got = %v, want %v", got, tt.want)
			}
		})
	}
}
