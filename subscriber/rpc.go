package subscriber

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/smartcontractkit/external-initiator/store"

	"github.com/smartcontractkit/chainlink/core/logger"
	"go.uber.org/atomic"
)

// RpcSubscriber holds the configuration for
// an RPC subscription.
type RpcSubscriber struct {
	Endpoint string
	Interval time.Duration

	client http.Client

	quitOnce sync.Once
	nonce    atomic.Uint64

	chClose chan struct{}
}

func NewRPCSubscriber(endpoint store.Endpoint) (*RpcSubscriber, error) {
	rpc := &RpcSubscriber{
		Endpoint: endpoint.Url,
		Interval: time.Duration(endpoint.RefreshInt) * time.Second,
		chClose:  make(chan struct{}),
	}

	return rpc, nil
}

func (rpc *RpcSubscriber) Type() Type {
	return RPC
}

func (rpc *RpcSubscriber) Stop() {
	rpc.quitOnce.Do(func() {
		close(rpc.chClose)
	})
}

func (rpc *RpcSubscriber) Subscribe(ctx context.Context, method, _ string, params json.RawMessage, ch chan<- json.RawMessage) error {
	// Make an initial request here to identify any issues
	res, err := rpc.Request(ctx, method, params)
	if err != nil {
		return err
	}

	// Launch the ticker goroutine after we've verified that the initial
	// request works.
	go func() {
		// Write to the channel in the goroutine here on purpose.
		// Since this action is blocking (for unbuffered ch), and the Subscribe function returns a potential
		// error, we can't assume that the reader is already running.
		ch <- res

		ticker := time.NewTicker(rpc.Interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				res, err := rpc.Request(ctx, method, params)
				if err != nil {
					logger.Error(err)
					continue
				}
				ch <- res
			case <-ctx.Done():
				return
			case <-rpc.chClose:
				return
			}
		}
	}()

	return nil
}

func (rpc *RpcSubscriber) Request(ctx context.Context, method string, params json.RawMessage) (json.RawMessage, error) {
	payload, err := NewJsonrpcMessage(rpc.nonce.Inc(), method, params)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, rpc.Endpoint, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := rpc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer logger.ErrorIfCalling(resp.Body.Close)

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var msg JsonrpcMessage
	return msg.Result, json.Unmarshal(respBody, &msg)
}
