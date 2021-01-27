package client

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/require"
)

type storeClientFailer struct {
	error        error
	closeError   error
	deleteError  error
	endpointName string
}

func (s storeClientFailer) DeleteAllEndpointsExcept([]string) error {
	return s.error
}

func (s storeClientFailer) LoadSubscriptions() ([]store.Subscription, error) {
	return nil, s.error
}

func (s storeClientFailer) LoadSubscription(string) (*store.Subscription, error) {
	return &store.Subscription{}, s.error
}

func (s storeClientFailer) LoadEndpoint(string) (store.Endpoint, error) {
	return store.Endpoint{Name: s.endpointName}, s.error
}

func (s storeClientFailer) Close() error {
	return s.closeError
}

func (s storeClientFailer) SaveSubscription(*store.Subscription) error {
	return s.error
}

func (s storeClientFailer) DeleteSubscription(*store.Subscription) error {
	return s.deleteError
}

func (s storeClientFailer) SaveEndpoint(*store.Endpoint) error {
	return s.error
}

func (s storeClientFailer) DB() *gorm.DB {
	return nil
}

type mockSubscription struct{}

func (s mockSubscription) Unsubscribe() {}

func Test_getSubscriber(t *testing.T) {
	ethWsManager, err := blockchain.CreateJsonManager(subscriber.WS, store.Subscription{
		Endpoint: store.Endpoint{
			Url:  "ws://localhost",
			Type: blockchain.ETH,
		},
	})
	require.NoError(t, err)

	ethRpcManager, err := blockchain.CreateJsonManager(subscriber.RPC, store.Subscription{
		Endpoint: store.Endpoint{
			Url:  "http://localhost",
			Type: blockchain.ETH,
		},
		Ethereum: store.EthSubscription{},
	})
	require.NoError(t, err)

	type args struct {
		sub store.Subscription
	}
	tests := []struct {
		name    string
		args    args
		want    subscriber.ISubscriber
		wantErr bool
	}{
		{
			"fails on invalid connection type",
			args{sub: store.Subscription{
				Endpoint: store.Endpoint{
					Url: "postgres://localhost",
				},
			}},
			nil,
			true,
		},
		{
			"fails on invalid subscription manager",
			args{sub: store.Subscription{
				Endpoint: store.Endpoint{
					Url: "http://localhost",
				},
			}},
			nil,
			true,
		},
		{
			"creates WS subscriber",
			args{sub: store.Subscription{
				Endpoint: store.Endpoint{
					Url:  "ws://localhost",
					Type: blockchain.ETH,
				},
			}},
			subscriber.WebsocketSubscriber{
				Endpoint: "ws://localhost",
				Manager:  ethWsManager,
			},
			false,
		},
		{
			"creates RPC subscriber",
			args{sub: store.Subscription{
				Endpoint: store.Endpoint{
					Url:        "http://localhost",
					Type:       blockchain.ETH,
					RefreshInt: 42,
				},
			}},
			subscriber.RpcSubscriber{
				Endpoint: "http://localhost",
				Interval: time.Duration(42) * time.Second,
				Manager:  ethRpcManager,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getSubscriber(tt.args.sub)
			if (err != nil) != tt.wantErr {
				t.Errorf("getSubscriber() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getSubscriber() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_normalizeLocalhost(t *testing.T) {
	type args struct {
		endpoint string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			"adds protocol when missing from localhost",
			args{"localhost"},
			"http://localhost",
		},
		{
			"doesn't add protocol to other domains",
			args{"chain.link"},
			"chain.link",
		},
		{
			"doesn't add protocol when already present",
			args{"http://localhost"},
			"http://localhost",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeLocalhost(tt.args.endpoint); got != tt.want {
				t.Errorf("normalizeLocalhost() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Service_DeleteJob(t *testing.T) {
	type fields struct {
		clNode        chainlink.Node
		store         storeInterface
		subscriptions map[string]*activeSubscription
	}
	type args struct {
		jobid string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"deletes sample job",
			fields{
				store: storeClientFailer{},
				subscriptions: map[string]*activeSubscription{
					"testJob": {
						Interface: mockSubscription{},
						Events:    make(chan subscriber.Event),
					},
				},
			},
			args{"testJob"},
			false,
		},
		{
			"deletes sample job even if not properly subscribed",
			fields{
				store: storeClientFailer{},
				subscriptions: map[string]*activeSubscription{
					"testJob": {},
				},
			},
			args{"testJob"},
			false,
		},
		{
			"deletes sample job even if not subscribed at all",
			fields{
				store:         storeClientFailer{},
				subscriptions: map[string]*activeSubscription{},
			},
			args{"testJob"},
			false,
		},
		{
			"fails on non-existent job",
			fields{
				store:         storeClientFailer{error: errors.New("record not found")},
				subscriptions: map[string]*activeSubscription{},
			},
			args{"testJob"},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Service{
				clNode:        tt.fields.clNode,
				store:         tt.fields.store,
				subscriptions: tt.fields.subscriptions,
			}
			if err := srv.DeleteJob(tt.args.jobid); (err != nil) != tt.wantErr {
				t.Errorf("DeleteJob() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_Service_GetEndpoint(t *testing.T) {
	type fields struct {
		clNode        chainlink.Node
		store         storeInterface
		subscriptions map[string]*activeSubscription
	}
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *store.Endpoint
		wantErr bool
	}{
		{
			"fetches the endpoint",
			fields{
				store: storeClientFailer{endpointName: "testEndpoint"},
			},
			args{"testEndpoint"},
			&store.Endpoint{
				Name: "testEndpoint",
			},
			false,
		},
		{
			"fails fetching non-existent endpoint",
			fields{
				store: storeClientFailer{error: errors.New("record not found")},
			},
			args{"testEndpoint"},
			nil,
			true,
		},
		{
			"fails with name mismatch",
			fields{
				store: storeClientFailer{endpointName: "wrongEndpoint"},
			},
			args{"testEndpoint"},
			nil,
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Service{
				clNode:        tt.fields.clNode,
				store:         tt.fields.store,
				subscriptions: tt.fields.subscriptions,
			}
			got, err := srv.GetEndpoint(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetEndpoint() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetEndpoint() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_Service_SaveEndpoint(t *testing.T) {
	type fields struct {
		clNode        chainlink.Node
		store         storeInterface
		subscriptions map[string]*activeSubscription
	}
	type args struct {
		e *store.Endpoint
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			"saves endpoint successfully",
			fields{
				store: storeClientFailer{},
			},
			args{e: &store.Endpoint{Name: "testEndpoint", Type: blockchain.ETH}},
			false,
		},
		{
			"fails endpoint validation",
			fields{
				store: storeClientFailer{},
			},
			args{e: &store.Endpoint{}},
			true,
		},
		{
			"fails save",
			fields{
				store: storeClientFailer{error: errors.New("could not save")},
			},
			args{e: &store.Endpoint{Name: "testEndpoint", Type: blockchain.ETH}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Service{
				clNode:        tt.fields.clNode,
				store:         tt.fields.store,
				subscriptions: tt.fields.subscriptions,
			}
			if err := srv.SaveEndpoint(tt.args.e); (err != nil) != tt.wantErr {
				t.Errorf("SaveEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_validateEndpoint(t *testing.T) {
	type args struct {
		endpoint store.Endpoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			"successfully validates bare-minimum endpoint",
			args{store.Endpoint{
				Type: blockchain.ETH,
				Name: "testEndpoint",
			}},
			false,
		},
		{
			"fails with invalid URL",
			args{store.Endpoint{
				Type: blockchain.ETH,
				Name: "testEndpoint",
				Url:  "http://a b.com/",
			}},
			true,
		},
		{
			"fails with invalid type",
			args{store.Endpoint{
				Type: "",
				Name: "testEndpoint",
			}},
			true,
		},
		{
			"fails with missing name",
			args{store.Endpoint{
				Type: blockchain.ETH,
				Name: "",
			}},
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateEndpoint(tt.args.endpoint); (err != nil) != tt.wantErr {
				t.Errorf("validateEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
