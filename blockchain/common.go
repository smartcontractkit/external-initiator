// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promLastSourcePing = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ei_last_source_ping",
		Help: "The timestamp of the last source of life from the source",
	}, []string{"endpoint", "jobid"})
)

var (
	ErrConnectionType = errors.New("unknown connection type")
	ErrSubscriberType = errors.New("unknown subscriber type")
)

// ExpectsMock variable is set when we run in a mock context
var ExpectsMock = false

var blockchains = []string{
	Substrate,
}

type Params struct {
	Endpoint    string          `json:"endpoint"`
	Addresses   []string        `json:"addresses"`
	Topics      []string        `json:"topics"`
	AccountIds  []string        `json:"accountIds"`
	Address     string          `json:"address"`
	UpkeepID    string          `json:"upkeepId"`
	ServiceName string          `json:"serviceName"`
	From        string          `json:"from"`
	FluxMonitor json.RawMessage `json:"fluxmonitor"`
}

// CreateJsonManager creates a new instance of a JSON blockchain manager with the provided
// connection type and store.Subscription config.
func CreateJsonManager(t subscriber.Type, sub store.Subscription) (subscriber.JsonManager, error) {
	switch sub.Endpoint.Type {
	case Substrate:
		// TODO: Implement
		return nil, nil
	}

	return nil, fmt.Errorf("unknown blockchain type %v for JSON manager", sub.Endpoint.Type)
}

// CreateClientManager creates a new instance of a subscriber.ISubscriber with the provided
// connection type and store.Subscription config.
func CreateClientManager(sub store.Subscription) (subscriber.ISubscriber, error) {
	switch sub.Endpoint.Type {
	}

	return nil, errors.New("unknown blockchain type for Client subscription")
}

func GetConnectionType(endpoint store.Endpoint) (subscriber.Type, error) {
	switch endpoint.Type {
	// Add blockchain implementations that encapsulate entire connection here
	case "": // TODO: XTZ, ONT, IOTX, Keeper, BIRITA:
		return subscriber.Client, nil
	default:
		u, err := url.Parse(endpoint.Url)
		if err != nil {
			return subscriber.Unknown, err
		}

		if strings.HasPrefix(u.Scheme, "ws") {
			return subscriber.WS, nil
		} else if strings.HasPrefix(u.Scheme, "http") {
			return subscriber.RPC, nil
		}

		return subscriber.Unknown, errors.New("unknown connection scheme")
	}
}

func ValidBlockchain(name string) bool {
	for _, blockchain := range blockchains {
		if name == blockchain {
			return true
		}
	}
	return false
}

func GetValidations(t string, params Params) []int {
	fmt.Println("T:")

	fmt.Println(t)
	switch t {
	case Substrate:
		return []int{
			len(params.AccountIds) + len(params.FluxMonitor),
		}
	}

	return nil
}

func CreateSubscription(sub *store.Subscription, params Params) {
	// fmConfig := services.ParseFMSpec(params.FluxMonitor)
	// go services.NewFluxMonitor(fmConfig)
	// FM probably needs to get started at createsubscription too. Check how to handle this.
	switch sub.Endpoint.Type {
	case Substrate:
		sub.Substrate = store.SubstrateSubscription{
			AccountIds: params.AccountIds,
		}
	}
}

// JsonrpcMessage declares JSON-RPC message type
type JsonrpcMessage struct {
	Version string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id,omitempty"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params,omitempty"`
	Error   *interface{}    `json:"error,omitempty"`
	Result  json.RawMessage `json:"result,omitempty"`
}

func convertStringArrayToKV(data []string) map[string]string {
	result := make(map[string]string)
	var key string

	for i, val := range data {
		if len(val) == 0 {
			continue
		}

		if i%2 == 0 {
			key = val
		} else if len(key) != 0 {
			result[key] = val
			key = ""
		}
	}

	return result
}

// matchesJobID checks if expected jobID matches the actual one, or are we in a mock context.
func matchesJobID(expected string, actual string) bool {
	if actual == expected {
		return true
	} else if ExpectsMock && actual == "mock" {
		return true
	}

	return false
}
