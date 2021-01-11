package client

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type storeFailer struct {
	error         error
	endpoint      *store.Endpoint
	endpointError error
}

func (s storeFailer) SaveSubscription(*store.Subscription) error {
	return s.error
}

func (s storeFailer) DeleteJob(string) error {
	return s.error
}

func (s storeFailer) GetEndpoint(string) (*store.Endpoint, error) {
	return s.endpoint, s.endpointError
}

func (s storeFailer) SaveEndpoint(*store.Endpoint) error {
	return s.error
}

func generateCreateSubscriptionReq(id, endpoint string, addresses, topics, accountIds []string) CreateSubscriptionReq {
	params := struct {
		Endpoint    string   `json:"endpoint"`
		Addresses   []string `json:"addresses"`
		Topics      []string `json:"topics"`
		AccountIds  []string `json:"accountIds"`
		Address     string   `json:"address"`
		UpkeepID    string   `json:"upkeepId"`
		ServiceName string   `json:"serviceName"`
	}{
		Endpoint:   endpoint,
		Addresses:  addresses,
		Topics:     topics,
		AccountIds: accountIds,
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
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil, &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusCreated,
		},
		{
			"Missing fields",
			generateCreateSubscriptionReq("id", "", []string{}, []string{}, []string{}),
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
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("failed save"), &store.Endpoint{Name: "eth-mainnet", Type: "ethereum"}, nil},
			http.StatusInternalServerError,
		},
		{
			"Endpoint does not exist",
			generateCreateSubscriptionReq("id", "doesnt-exist", []string{"0x123"}, []string{"0x123"}, []string{"0x123"}),
			storeFailer{errors.New("Failed loading endpoint"), nil, nil},
			http.StatusBadRequest,
		},
		{
			"Failed fetching Endpoint",
			generateCreateSubscriptionReq("id", "eth-mainnet", []string{"0x123"}, []string{"0x123"}, []string{"0x123"}),
			storeFailer{nil, nil, errors.New("failed SQL query")},
			http.StatusInternalServerError,
		},
	}
	for _, test := range tests {
		t.Log(test.Name)
		body, err := json.Marshal(test.Payload)
		require.NoError(t, err)

		srv := &HttpService{
			Store: test.App,
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
		srv := &HttpService{
			Store: test.App,
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
		srv := &HttpService{}
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

func TestRequireAuth(t *testing.T) {
	key := "testKey"
	secret := "testSecretAbcdæøå"

	tests := []struct {
		Name   string
		Method string
		Target string
		Auth   bool
	}{
		{
			"Health is open",
			"GET",
			"/health",
			false,
		},
		{
			"Creating jobs is protected",
			"POST",
			"/jobs",
			true,
		},
		{
			"Deleting jobs is protected",
			"DELETE",
			"/jobs/test",
			true,
		},
		{
			"Creating config is protected",
			"POST",
			"/config",
			true,
		},
	}

	srv := &HttpService{
		AccessKey: key,
		Secret:    secret,
		Store:     storeFailer{error: errors.New("testing only")},
	}
	srv.createRouter()

	for _, test := range tests {
		req := httptest.NewRequest(test.Method, test.Target, nil)

		w := httptest.NewRecorder()
		srv.ServeHTTP(w, req)
		if test.Auth {
			assert.Equal(t, http.StatusUnauthorized, w.Code)

			req.Header.Set(externalInitiatorAccessKeyHeader, key)
			req.Header.Set(externalInitiatorSecretHeader, secret)

			w = httptest.NewRecorder()
			srv.ServeHTTP(w, req)
			assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		} else {
			assert.NotEqual(t, http.StatusUnauthorized, w.Code)
		}
	}
}

func Test_httpService_CreateEndpoint(t *testing.T) {
	tests := []struct {
		Name       string
		Payload    interface{}
		App        subscriptionStorer
		StatusCode int
	}{
		{
			"Create success",
			store.Endpoint{
				Url:  "http://localhost/",
				Name: "test",
			},
			storeFailer{},
			http.StatusCreated,
		},
		{
			"Missing fields",
			store.Endpoint{
				Url: "http://localhost/",
			},
			storeFailer{error: errors.New("missing name")},
			http.StatusInternalServerError,
		},
		{
			"Decode failed",
			"bad json format",
			storeFailer{},
			http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Log(test.Name)
		body, err := json.Marshal(test.Payload)
		require.NoError(t, err)

		srv := &HttpService{
			Store: test.App,
		}
		srv.createRouter()

		req := httptest.NewRequest("POST", "/config", bytes.NewBuffer(body))

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
