package harmony

import (
	"testing"

	"github.com/smartcontractkit/external-initiator/store"

	"github.com/stretchr/testify/require"
)

func Test_createManager(t *testing.T) {
	rpcSub := store.Subscription{Job: "1234-rpc", Endpoint: store.Endpoint{Url: "https://example.com"}}
	rpcManager, err := createManager(rpcSub)
	require.NoError(t, err)

	type args struct {
		sub store.Subscription
	}
	tests := []struct {
		name    string
		args    args
		want    *manager
		wantErr bool
	}{
		{
			"creates manager with RPC subscriber",
			args{rpcSub},
			rpcManager,
			false,
		},
		{
			"fails with invalid URL scheme",
			args{store.Subscription{Job: "1234-invalid-url", Endpoint: store.Endpoint{Url: "not valid"}}},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createManager(tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("createManager() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("createManager() got nil")
				return
			}
			if got != nil && tt.want.subscriber.Type() != got.subscriber.Type() {
				t.Errorf("createManager() got Type() = %v, want %v", got.subscriber.Type(), tt.want.subscriber.Type())
			}
		})
	}
}
