package blockchain

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/irisnet/service-sdk-go/service"
	"github.com/irisnet/service-sdk-go/types"

	"github.com/smartcontractkit/external-initiator/store"
)

func TestCreateBSNIritaSubscriber(t *testing.T) {
	t.Run("creates biritaSubscriber from subscription",
		func(t *testing.T) {
			sub := store.Subscription{
				Job: "test",
				BSNIrita: store.BSNIritaSubscription{
					Addresses:   []string{"test-provider-address"},
					ServiceName: "oracle",
				},
			}
			biritaSubscriber := createBSNIritaSubscriber(sub)
			assert.Equal(t, "oracle", biritaSubscriber.ServiceName)
			assert.Equal(t, []string{"test-provider-address"}, biritaSubscriber.Addresses)
		})
}

func TestBuildTriggerEvent(t *testing.T) {
	providerAddrBech32 := "iaa1cq3xx80jym3jshlxmwnfwu840jxta032aa4jss"
	providerAddr, _ := types.AccAddressFromBech32(providerAddrBech32)
	requestID, _ := hex.DecodeString("FFB2EA8819BAF485C49DEBC08A4431E4BA5707945F8B33C8E777110BE62491240000000000000000000000000000007900000000000017F70000")

	tests := []struct {
		name     string
		args     service.Request
		wantPass bool
		want     []byte
	}{
		{
			"basic service request",
			service.Request{
				Id:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			true,
			[]byte(`{"request_id":"FFB2EA8819BAF485C49DEBC08A4431E4BA5707945F8B33C8E777110BE62491240000000000000000000000000000007900000000000017F70000","request_body":"{\"id\":\"test\"}"}`),
		},
		{
			"missing request id",
			service.Request{
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"missing request input",
			service.Request{
				Id:          requestID,
				Input:       "",
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"service name does not match",
			service.Request{
				Id:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "incorrect-service-name",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"provider address does not match",
			service.Request{
				Id:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    types.AccAddress([]byte("incorrect-provider")),
			},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &biritaSubscription{
				addresses:   map[string]bool{providerAddrBech32: true},
				serviceName: "oracle",
			}

			event, err := sub.buildTriggerEvent(tt.args)
			if tt.wantPass {
				assert.NoError(t, err, "buildTriggerEvent not passed, wantPass %v", tt.wantPass)
			} else {
				assert.Error(t, err, "buildTriggerEvent passed, wantPass %v", tt.wantPass)
			}

			if !bytes.Equal(event, tt.want) {
				t.Errorf("buildTriggerEvent got = %s, want %s", event, tt.want)
			}
		})
	}
}
