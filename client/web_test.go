package client

import (
	"bytes"
	"encoding/json"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

type storeFailer struct{ error }

func (s storeFailer) SaveSubscription(arg *store.Subscription) error {
	return s.error
}

func generateCreateSubscriptionReq(id, chain, endpoint string, addresses, topics []string) CreateSubscriptionReq {
	params := struct {
		Type      string   `json:"type"`
		Endpoint  string   `json:"endpoint"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"initiatorTopics"`
	}{
		Type:      chain,
		Endpoint:  endpoint,
		Addresses: addresses,
		Topics:    topics,
	}

	return CreateSubscriptionReq{
		JobID:  id,
		Type:   "external",
		Params: params,
	}
}

func TestConfigController(t *testing.T) {
	tests := []struct {
		Name       string
		Payload    interface{}
		App        subscriptionStorer
		StatusCode int
	}{
		{
			"Create success",
			generateCreateSubscriptionReq("id", "ethereum", "eth-mainnet", []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil},
			http.StatusCreated,
		},
		{
			"Missing fields",
			generateCreateSubscriptionReq("id", "", "", []string{}, []string{}),
			storeFailer{nil},
			http.StatusBadRequest,
		},
		{
			"Decode failed",
			"bad json format",
			storeFailer{errors.New("failed save")},
			http.StatusBadRequest,
		},
		{
			"Save failed",
			generateCreateSubscriptionReq("id", "ethereum", "eth-mainnet", []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("failed save")},
			http.StatusInternalServerError,
		},
		{
			"Endpoint does not exist",
			generateCreateSubscriptionReq("id", "ethereum", "doesnt-exist", []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("Failed loading endpoint")},
			http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Log(test.Name)
		body, err := json.Marshal(test.Payload)
		require.NoError(t, err)

		srv := &httpService{
			store: test.App,
		}
		srv.createRouter()

		req := httptest.NewRequest("POST", "/job", bytes.NewBuffer(body))

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		var respJSON map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &respJSON)
		assert.NoError(t, err)
	}
}

func TestHealthController(t *testing.T) {
	tests := []struct {
		Name       string
		StatusCode int
	}{
		{
			"Is healthy",
			http.StatusOK,
		},
	}
	for _, test := range tests {
		srv := &httpService{}
		srv.createRouter()

		req := httptest.NewRequest("GET", "/health", nil)

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		var respJSON map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &respJSON)
		assert.NoError(t, err)
	}
}
