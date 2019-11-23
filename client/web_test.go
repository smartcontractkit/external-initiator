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

type storeFailer struct {
	error         error
	endpoint      *store.Endpoint
	endpointError error
}

func (s storeFailer) SaveSubscription(arg *store.Subscription) error {
	return s.error
}

func (s storeFailer) DeleteJob(jobid string) error {
	return s.error
}

func (s storeFailer) GetEndpoint(name string) (*store.Endpoint, error) {
	return s.endpoint, s.endpointError
}

func (s storeFailer) SaveEndpoint(e *store.Endpoint) error {
	return s.error
}

func generateCreateSubscriptionReq(id, endpoint string, addresses, topics []string) CreateSubscriptionReq {
	params := struct {
		Endpoint  string   `json:"endpoint"`
		Addresses []string `json:"addresses"`
		Topics    []string `json:"eventTopics"`
	}{
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
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil, &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusCreated,
		},
		{
			"Missing fields",
			generateCreateSubscriptionReq("id", "", []string{}, []string{}),
			storeFailer{nil, &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusBadRequest,
		},
		{
			"Decode failed",
			"bad json format",
			storeFailer{errors.New("failed save"), &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusBadRequest,
		},
		{
			"Save failed",
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("failed save"), &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusInternalServerError,
		},
		{
			"Endpoint does not exist",
			generateCreateSubscriptionReq("id", "doesnt-exist", []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("Failed loading endpoint"), nil, nil},
			http.StatusBadRequest,
		},
		{
			"Failed fetching Endpoint",
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil, nil, errors.New("failed SQL query")},
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

		req := httptest.NewRequest("POST", "/jobs", bytes.NewBuffer(body))

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		if w.Code == http.StatusNotFound {
			// Do not expect JSON response
			continue
		}

		var respJSON map[string]interface{}
		err = json.Unmarshal(w.Body.Bytes(), &respJSON)
		assert.NoError(t, err)
	}
}

func TestDeleteController(t *testing.T) {
	tests := []struct {
		Name       string
		Jobid      string
		App        subscriptionStorer
		StatusCode int
	}{
		{
			"Delete success",
			"test",
			storeFailer{nil, nil, nil},
			http.StatusOK,
		},
		{
			"Missing jobid",
			"",
			storeFailer{errors.New("missing jobid"), nil, nil},
			http.StatusNotFound,
		},
		{
			"Failed deleting job",
			"test",
			storeFailer{errors.New("record not found"), nil, nil},
			http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Log(test.Name)
		srv := &httpService{
			store: test.App,
		}
		srv.createRouter()

		req := httptest.NewRequest("DELETE", "/jobs/"+test.Jobid, nil)

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		assert.Equal(t, test.StatusCode, w.Code)

		if w.Code == http.StatusNotFound {
			// Do not expect JSON response
			continue
		}

		var respJSON map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &respJSON)
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
