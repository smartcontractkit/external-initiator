package client

import (
	"errors"
	"github.com/smartcontractkit/external-initiator/blockchain/substrate"
	"reflect"
	"testing"

	"github.com/smartcontractkit/external-initiator/chainlink"
	"github.com/smartcontractkit/external-initiator/store"
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

func (s storeClientFailer) SaveJobSpec(*store.JobSpec) error {
	return s.error
}

func (s storeClientFailer) LoadJobSpec(string) (*store.JobSpec, error) {
	return nil, s.error
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
					"testJob": {},
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
			args{e: &store.Endpoint{Name: "testEndpoint", Type: substrate.Name}},
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
			args{e: &store.Endpoint{Name: "testEndpoint", Type: substrate.Name}},
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
				Type: substrate.Name,
				Name: "testEndpoint",
			}},
			false,
		},
		{
			"fails with invalid URL",
			args{store.Endpoint{
				Type: substrate.Name,
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
				Type: substrate.Name,
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
