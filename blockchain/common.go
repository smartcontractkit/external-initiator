// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"encoding/json"
	"errors"
	"net/url"
	"strings"

	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

var ExpectsMock = false

var blockchains = []string{
	ETH,
	HMY,
	XTZ,
	Substrate,
	ONT,
	BSC,
	NEAR,
}

type Params struct {
	Endpoint   string   `json:"endpoint"`
	Addresses  []string `json:"addresses"`
	Topics     []string `json:"eventTopics"`
	AccountIds []string `json:"accountIds"`
}

// CreateJsonManager creates a new instance of a JSON blockchain manager with the provided
// connection type and store.Subscription config.
func CreateJsonManager(t subscriber.Type, sub store.Subscription) (subscriber.JsonManager, error) {
	switch sub.Endpoint.Type {
	case ETH:
		return createEthManager(t, sub), nil
	case HMY:
		return createHmyManager(t, sub), nil
	case Substrate:
		return createSubstrateManager(t, sub)
	case BSC:
		return createBscManager(t, sub), nil
	case NEAR:
		return createNEARManager(t, sub)
	}

	return nil, errors.New("unknown blockchain type for JSON manager")
}

// CreateClientManager creates a new instance of a subscriber.ISubscriber with the provided
// connection type and store.Subscription config.
func CreateClientManager(sub store.Subscription) (subscriber.ISubscriber, error) {
	switch sub.Endpoint.Type {
	case XTZ:
		return createTezosSubscriber(sub), nil
	case ONT:
		return createOntSubscriber(sub), nil
	}

	return nil, errors.New("unknown blockchain type for Client subscription")
}

func GetConnectionType(endpoint store.Endpoint) (subscriber.Type, error) {
	switch endpoint.Type {
	// Add blockchain implementations that encapsulate entire connection here
	case XTZ, ONT:
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
	switch t {
	case ETH, HMY:
		return []int{
			len(params.Addresses) + len(params.Topics),
		}
	case XTZ:
		return []int{
			len(params.Addresses),
		}
	case Substrate:
		return []int{
			len(params.AccountIds),
		}
	case ONT:
		return []int{
			len(params.Addresses),
		}
	case BSC:
		return []int{
			len(params.Addresses),
		}
	case NEAR:
		return []int{
			len(params.Addresses),
		}
	}

	return nil
}

func CreateSubscription(sub *store.Subscription, params Params) {
	switch sub.Endpoint.Type {
	case ETH, HMY:
		sub.Ethereum = store.EthSubscription{
			Addresses: params.Addresses,
			Topics:    params.Topics,
		}
	case XTZ:
		sub.Tezos = store.TezosSubscription{
			Addresses: params.Addresses,
		}
	case Substrate:
		sub.Substrate = store.SubstrateSubscription{
			AccountIds: params.AccountIds,
		}
	case ONT:
		sub.Ontology = store.OntSubscription{
			Addresses: params.Addresses,
		}
	case BSC:
		sub.BinanceSmartChain = store.BinanceSmartChainSubscription{
			Addresses: params.Addresses,
		}
	case NEAR:
		sub.NEAR = store.NEARSubscription{
			AccountIds: params.Addresses,
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
