package blockchain

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

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
			biritaSubscriber, err := createBSNIritaSubscriber(sub)
			assert.NoError(t, err)
			assert.Equal(t, "oracle", biritaSubscriber.ServiceName)
			assert.Equal(t, []string{"test-provider-address"}, biritaSubscriber.Addresses)
		})
}

func TestBuildTriggerEvent(t *testing.T) {
	providerAddr := "iaa1cq3xx80jym3jshlxmwnfwu840jxta032aa4jss"
	requestID := "FFB2EA8819BAF485C49DEBC08A4431E4BA5707945F8B33C8E777110BE62491240000000000000000000000000000007900000000000017F70000"

	tests := []struct {
		name     string
		args     BIritaServiceRequest
		wantPass bool
		want     []byte
	}{
		{
			"basic service request",
			BIritaServiceRequest{
				ID:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			true,
			[]byte(`{"request_id":"FFB2EA8819BAF485C49DEBC08A4431E4BA5707945F8B33C8E777110BE62491240000000000000000000000000000007900000000000017F70000","request_body":{"id":"test"}}`),
		},
		{
			"missing request id",
			BIritaServiceRequest{
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"missing request input",
			BIritaServiceRequest{
				ID:          requestID,
				Input:       "",
				ServiceName: "oracle",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"service name does not match",
			BIritaServiceRequest{
				ID:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "incorrect-service-name",
				Provider:    providerAddr,
			},
			false,
			nil,
		},
		{
			"provider address does not match",
			BIritaServiceRequest{
				ID:          requestID,
				Input:       `{"body":{"id":"test"}}`,
				ServiceName: "oracle",
				Provider:    "incorrect-provider",
			},
			false,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sub := &biritaSubscription{
				addresses:   map[string]bool{providerAddr: true},
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
