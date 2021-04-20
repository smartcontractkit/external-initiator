// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"fmt"

	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/blockchain/ethereum"
	"github.com/smartcontractkit/external-initiator/blockchain/substrate"
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

func CreateFluxMonitorManager(sub store.Subscription) (common.FluxMonitorManager, error) {
	switch sub.Endpoint.Type {
	case substrate.Name:
		return substrate.CreateFluxMonitorManager(sub)
	}
	return nil, fmt.Errorf("unknown endpoint type: %s", sub.Endpoint.Type)
}

func CreateRunlogManager(sub store.Subscription) (common.RunlogManager, error) {
	switch sub.Endpoint.Type {
	case ethereum.Name:
		return ethereum.CreateRunlogManager(sub)
	case substrate.Name:
		return substrate.CreateRunlogManager(sub)
	}
	return nil, fmt.Errorf("unknown endpoint type: %s", sub.Endpoint.Type)
}

var blockchains = []string{
	substrate.Name,
	ethereum.Name,
}

func ValidBlockchain(name string) bool {
	for _, blockchain := range blockchains {
		if name == blockchain {
			return true
		}
	}
	return false
}

func ValidateParams(t string, params common.Params) (missing []string) {
	switch t {
	case substrate.Name:
		if params.FluxMonitor == nil {
			return
		}
		if params.AccountId == "" {
			missing = append(missing, "account_id")
		}
		if params.FeedId == nil {
			missing = append(missing, "feed_id")
		}
	}

	return
}

func CreateSubscription(sub store.Subscription, params common.Params) store.Subscription {
	addresses := params.Addresses
	if len(params.Addresses) == 0 && params.Address != "" {
		addresses = []string{params.Address}
	}

	switch sub.Endpoint.Type {
	case ethereum.Name:
		sub.Ethereum = store.EthSubscription{
			Addresses: addresses,
		}
	case substrate.Name:
		sub.Substrate = store.SubstrateSubscription{
			AccountIds: params.AccountIds,
			FeedId:     *params.FeedId,
			AccountId:  params.AccountId,
		}
	}

	return sub
}
