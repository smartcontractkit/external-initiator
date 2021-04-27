package harmony

import (
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Name = "harmony"

// The manager implements the subscriber.JsonManager interface and allows
// for interacting with ETH nodes over RPC or WS.
type manager struct {
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
		endpointName: sub.EndpointName,
		jobid:        sub.Job,
		subscriber:   conn,
	}, nil
}

func (rm runlogManager) Stop() {
	// TODO: Implement me
}
