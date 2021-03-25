// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"encoding/json"
	"errors"
	"math/big"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shopspring/decimal"
	"github.com/smartcontractkit/external-initiator/store"
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

const (
	FMRequestState    = "fm_requestState"
	FMSubscribeEvents = "fm_subscribeEvents"
	FMJobRun          = "fm_jobRun"
)

type FluxAggregatorState struct {
	CurrentRoundID *int32
	LatestAnswer   *decimal.Decimal
	MinSubmission  *decimal.Decimal
	MaxSubmission  *decimal.Decimal
	Payment        *big.Int
	Timeout        *uint32
	RestartDelay   *int32
	//not sure if needed
	// LatestRoundID int32
	CanSubmit *bool

	OracleStarted *bool
}

type Manager interface {
	Request(t string) (response interface{}, err error)
	Subscribe(t string, ch chan<- interface{}) (err error)
	CreateJobRun(t string, params interface{}) (map[string]interface{}, error)
}

func CreateManager(sub store.Subscription) (Manager, error) {
	switch sub.Endpoint.Type {
	case Substrate:
		return createSubstrateManager(sub)
	}

	return nil, nil
}

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

	// Substrate FM:
	FeedId    uint32 `json:"feed_id"`
	AccountId string `json:"account_id"`
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
			FeedId:     params.FeedId,
			AccountId:  params.AccountId,
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
