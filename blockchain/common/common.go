// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package common

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/pkg/errors"
)

var (
	ErrConnectionType = errors.New("unknown connection type")
	ErrSubscriberType = errors.New("unknown subscriber type")
)

const (
	FMRequestState    = "fm_requestState"
	FMSubscribeEvents = "fm_subscribeEvents"
	FMJobRun          = "fm_jobRun"

	RunlogBackfill  = "runlog_backfill"
	RunlogSubscribe = "runlog_subscribe"
	RunlogJobRun    = "runlog_jobRun"
)

type FMEventNewRound struct {
	RoundID         uint32
	OracleInitiated bool
}

type FMEventAnswerUpdated struct {
	LatestAnswer big.Int
}

type FMEventPermissionsUpdated struct {
	CanSubmit bool
}

type FluxAggregatorState struct {
	RoundID       uint32
	LatestAnswer  big.Int
	MinSubmission big.Int
	MaxSubmission big.Int
	Payment       big.Int
	Timeout       uint32
	RestartDelay  int32
	CanSubmit     bool
}

type RunlogRequest struct {
	Params           map[string]string
	Payment          string
	RequestId        string
	CallbackFunction string
}

type Manager interface {
	Request(t string) (response interface{}, err error)
	Subscribe(ctx context.Context, t string, ch chan<- interface{}) (err error)
	CreateJobRun(t string, v interface{}) (map[string]interface{}, error)
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

	// Name FM:
	FeedId    uint32 `json:"feed_id"`
	AccountId string `json:"account_id"`
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

func ConvertStringArrayToKV(data []string) map[string]string {
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

// ExpectsMock variable is set when we run in a mock context
var ExpectsMock = false

// MatchesJobID checks if expected jobID matches the actual one, or are we in a mock context.
func MatchesJobID(expected string, actual string) bool {
	if actual == expected {
		return true
	} else if ExpectsMock && actual == "mock" {
		return true
	}

	return false
}
