// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"encoding/json"
	"fmt"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/blockchain/terra"
	"github.com/smartcontractkit/external-initiator/store"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	promLastSourcePing = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ei_last_source_ping",
		Help: "The timestamp of the last source of life from the source",
	}, []string{"endpoint", "jobid"})
)

type Params struct {
	Endpoint    string          `json:"endpoint"`
	UpkeepID    string          `json:"upkeepId"`
	ServiceName string          `json:"serviceName"`
	From        string          `json:"from"`
	FluxMonitor json.RawMessage `json:"fluxmonitor,omitempty"`

	terra.TerraParams
}

func CreateFluxMonitorManager(sub store.Subscription) (common.FluxMonitorManager, error) {
	switch sub.Endpoint.Type {
	case terra.Name:
		return terra.CreateFluxMonitorManager(sub)
	}
	return nil, fmt.Errorf("unknown endpoint type: %s", sub.Endpoint.Type)
}

var blockchains = []string{
	terra.Name,
}

func ValidBlockchain(name string) bool {
	for _, blockchain := range blockchains {
		if name == blockchain {
			return true
		}
	}
	return false
}

func ValidateParams(t string, params Params) (missing []string) {
	// switch t {
	// case ethereum.Name, conflux.Name, harmony.Name:
	// 	if len(params.Address)+len(params.Addresses) == 0 {
	// 		missing = append(missing, "address")
	// 	}
	// case substrate.Name:
	// 	if params.FluxMonitor == nil {
	// 		return
	// 	}
	// 	if params.AccountId == "" {
	// 		missing = append(missing, "account_id")
	// 	}
	// 	if params.FeedId == nil {
	// 		missing = append(missing, "feed_id")
	// 	}
	// }

	return
}

func CreateSubscription(sub store.Subscription, params Params) store.Subscription {
	switch sub.Endpoint.Type {
	case terra.Name:
		sub.Terra = store.TerraSubscription{
			ContractAddress: params.ContractAddress,
			AccountAddress:  params.AccountAddress,
		}
	}

	return sub
}
