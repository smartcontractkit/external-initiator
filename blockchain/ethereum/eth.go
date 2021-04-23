package ethereum

import (
	"github.com/smartcontractkit/external-initiator/blockchain/evm"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Name = "ethereum"

type EthParams struct {
	Address   string   `json:"address"`
	Addresses []string `json:"addresses"`
	Topics    []string `json:"topics"`
}

// The manager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type manager struct {
	fq           *evm.FilterQuery
	endpointName string
	jobid        string

	subscriber subscriber.ISubscriber
}

func createManager(sub store.Subscription) (*manager, error) {
	conn, err := subscriber.NewSubscriber(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		fq:           evm.CreateEvmFilterQuery(sub.Job, sub.Ethereum.Addresses),
		endpointName: sub.EndpointName,
		jobid:        sub.Job,
		subscriber:   conn,
	}, nil
}

func (rm runlogManager) Stop() {
	// TODO: Implement me
}
