// Package blockchain provides functionality to interact with
// different blockchain interfaces.
package blockchain

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/smartcontractkit/external-initiator/blockchain/common"
	"github.com/smartcontractkit/external-initiator/blockchain/substrate"
	"github.com/smartcontractkit/external-initiator/store"
)

var (
	promLastSourcePing = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "ei_last_source_ping",
		Help: "The timestamp of the last source of life from the source",
	}, []string{"endpoint", "jobid"})
)

func CreateManager(sub store.Subscription) (common.Manager, error) {
	switch sub.Endpoint.Type {
	case substrate.Name:
		return substrate.CreateSubstrateManager(sub)
	}
	return nil, fmt.Errorf("unknown endpoint type: %s", sub.Endpoint.Type)
}

var blockchains = []string{
	substrate.Name,
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
	switch sub.Endpoint.Type {
	case substrate.Name:
		sub.Substrate = store.SubstrateSubscription{
			AccountIds: params.AccountIds,
			FeedId:     *params.FeedId,
			AccountId:  params.AccountId,
		}
	}

	return sub
}
