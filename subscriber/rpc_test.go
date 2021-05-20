package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

var interval = 2
var testMethod = "true"
var emptyBytes = []byte("{}")

func getRPC(u url.URL) (*RpcSubscriber, error) {
	endpoint := getEndpoint(u)
	endpoint.RefreshInt = interval
	return NewRPCSubscriber(endpoint)
}

func TestNewRPCSubscriber(t *testing.T) {
	t.Run("creates new RPC subscriber", func(t *testing.T) {
		u := *rpcMockUrl

		// setup new subscriber
		rpc, err := getRPC(u)
		if err != nil {
			t.Errorf("NewRPCSubscriber() error = %v", err)
			return
		}
		defer rpc.Stop()

		// checking return parameters
		assert.Equal(t, u.String(), rpc.Endpoint)
		assert.Equal(t, fmt.Sprintf("%d%s", interval, "s"), rpc.Interval.String())
		assert.Equal(t, RPC, rpc.Type())
	})
}

func TestRPCRequest(t *testing.T) {
	t.Run("succeeds with normal request", func(t *testing.T) {
		u := *rpcMockUrl

		rpc, err := getRPC(u)
		if err != nil {
			t.Errorf("Request() error = %v", err)
			return
		}
		defer rpc.Stop()

		_, err = rpc.Request(context.TODO(), testMethod, emptyBytes)
		if err != nil {
			t.Errorf("Request() unexpected error = %v", err)
			return
		}

	})

	t.Run("fails on invalid endpoint", func(t *testing.T) {
		u, _ := url.Parse("http://localhost:8080")

		rpc, err := getRPC(*u)
		if err != nil {
			t.Errorf("Request() error (invalid) = %v", err)
			return
		}
		defer rpc.Stop()

		_, err = rpc.Request(context.TODO(), testMethod, emptyBytes)
		if err == nil {
			t.Errorf("Request() expects error (invalid), but got nil")
			return
		}
	})

	t.Run("fails on bad status", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "/fails"

		rpc, err := getRPC(u)
		if err != nil {
			t.Errorf("Request() error (bad status) = %v", err)
			return
		}
		defer rpc.Stop()

		_, err = rpc.Request(context.TODO(), testMethod, emptyBytes)
		if err == nil {
			t.Errorf("Request() expects error (bad status), but got nil")
			return
		}
	})
}

func TestRPCSubscribe(t *testing.T) {
	t.Run("subscribes to rpc endpoint", func(t *testing.T) {
		u := *rpcMockUrl
		u.Path = "test/subscribe"

		rpc, err := getRPC(u)
		if err != nil {
			t.Errorf("Subscribe() error (getRPC) = %v", err)
			return
		}
		defer rpc.Stop()

		listener := make(chan json.RawMessage)
		err = rpc.Subscribe(context.TODO(), testMethod, "", emptyBytes, listener)
		if err != nil {
			t.Errorf("Subscribe() error = %v", err)
			return
		}

		res := <-listener
		resStr := string(res)
		if resStr != "1" {
			t.Errorf("Subscribe() got unexpected first message = %v", resStr)
			return
		}

		res = <-listener
		resStr = string(res)
		if resStr != "2" {
			t.Errorf("Subscribe() got unexpected second message = %v", resStr)
			return
		}

	})
}
