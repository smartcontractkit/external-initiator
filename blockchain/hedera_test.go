package blockchain

import (
	"errors"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/assert"
	"reflect"
	"testing"
)

func init() {
	tokenId = "0.0.2138566"
	minPayment = 1000000000
}

func TestHedera_CreateHederaSubscriber(t *testing.T) {
	tests := []struct {
		name string
		args store.Subscription
		want hederaSubscriber
	}{
		{
			"empty",
			store.Subscription{},
			hederaSubscriber{},
		},
		{
			"enpoint only",
			store.Subscription{
				Endpoint: store.Endpoint{Url: "http://example.com/api"},
			},
			hederaSubscriber{Endpoint: "http://example.com/api"},
		},
		{
			"endpoint and accountId only",
			store.Subscription{
				Endpoint: store.Endpoint{Url: "http://example.com/api"},
				Hedera:   store.HederaSubscription{AccountId: "0.0.1234"},
			},
			hederaSubscriber{
				Endpoint:  "http://example.com/api",
				AccountId: "0.0.1234",
			},
		},
		{
			"endpoint and accountId and jobId",
			store.Subscription{
				Endpoint: store.Endpoint{Url: "http://example.com/api"},
				Hedera:   store.HederaSubscription{AccountId: "0.0.1234"},
				Job:      "8fb39e9048844fe58de078d46d8bc9d0"},
			hederaSubscriber{
				Endpoint:  "http://example.com/api",
				AccountId: "0.0.1234",
				JobID:     "8fb39e9048844fe58de078d46d8bc9d0",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := createHederaSubscriber(tt.args)
			if !reflect.DeepEqual(sub, tt.want) {
				t.Errorf("createHederaSubscriber() = %s, want %s", sub, tt.want)
			}
		})
	}
}

func TestHedera_DecodeMemo(t *testing.T) {
	tests := []struct {
		name string
		arg  string
		want string
	}{
		{
			"empty",
			"",
			"",
		},
		{
			"decode topic id and job id",
			"MC4wLjIxMTM4MTQtZnR2cGEgNTMyZmFiOTA4ZTJhNDg3MGI3ZDUwZTI4ZWViMGEzMjU=",
			"0.0.2113814-ftvpa 532fab908e2a4870b7d50e28eeb0a325",
		},
		{
			"decode incorrect memo",
			"-1",
			"",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := DecodeMemo(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("DecodeMemo(string) = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestHedera_FromString(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    int64
		wantErr error
	}{
		{
			"empty",
			"",
			0,
			errors.New("invalid timestamp seconds provided"),
		},
		{
			"empty seconds",
			".033488000",
			0,
			errors.New("invalid timestamp seconds provided"),
		},
		{
			"empty nanoseconds",
			"1626699754.",
			0,
			errors.New("invalid timestamp nanos provided"),
		},
		{
			"correct timestamp",
			"1626699754.033488000",
			1626699754033488000,
			nil,
		},
		{
			"incomplete nanoseconds",
			"1626699754.0",
			1626699754000000000,
			nil,
		},
		{
			"incomplete seconds",
			"0.033488000",
			33488000,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FromString(tt.arg)

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FromString(int64) got = %d, want %d", got, tt.want)
			}
			if err != nil && !assert.EqualError(t, err, tt.wantErr.Error()) {
				t.Errorf("Error should be: %v, got: %v", tt.wantErr, err)
			}
		})
	}
}

func TestHedera_String(t *testing.T) {
	tests := []struct {
		name string
		arg  int64
		want string
	}{
		{
			"zero",
			0,
			"0.0",
		},
		{
			"zero seconds",
			33488000,
			"0.33488000",
		},
		{
			"correct timestamp",
			1626699754033488000,
			"1626699754.33488000",
		},
		{
			"short timestamp",
			1626699754,
			"1.626699754",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := String(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("String(string) = %s, want %s", got, tt.want)
			}
		})
	}
}

func TestHedera_CheckForValidTokenTransfer(t *testing.T) {
	tests := []struct {
		name string
		arg1 []Transfer
		arg2 string
		want bool
	}{
		{
			"valid transfer",
			[]Transfer{
				{
					Token:   "0.0.2138566",
					Amount:  2500000000,
					Account: "0.0.1967249",
				},
				{
					Token:   "0.0.2138569",
					Amount:  2500000000,
					Account: "0.0.1967249",
				},
			},
			"0.0.1967249",
			true,
		},
		{
			"invalid transfer",
			[]Transfer{
				{
					Token:   "0.0.2138569",
					Amount:  2500000000,
					Account: "0.0.1967249",
				},
				{
					Token:   "0.0.2138569",
					Amount:  2500000000,
					Account: "0.0.1967249",
				},
			},
			"0.0.1967249",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			if got := checkForValidTokenTransfer(tt.arg1, tt.arg2); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("checkForValidTokenTransfer([]Transfer, string) = %t, want %t", got, tt.want)
			}
		})
	}
}
