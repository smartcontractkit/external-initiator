package ethereum

import (
	"github.com/smartcontractkit/external-initiator/blockchain/evm"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Name = "ethereum"

// The manager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type manager struct {
	fq           *evm.FilterQuery
	endpointName string
	jobid        string

	subscriber subscriber.ISubscriber
}

func createManager(sub store.Subscription) (*manager, error) {
	return &manager{
		fq:           evm.CreateEvmFilterQuery(sub.Job, sub.Ethereum.Addresses),
		endpointName: sub.EndpointName,
		jobid:        sub.Job,
	}, nil
}

func (rm runlogManager) Stop() {
	// TODO: Implement me
}
