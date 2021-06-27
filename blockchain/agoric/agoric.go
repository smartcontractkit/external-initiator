package agoric

import (
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

const Name = "agoric"

type manager struct {
	jobid string

	conn *subscriber.WebsocketConnection
}

func createManager(sub store.Subscription) (*manager, error) {
	conn, err := subscriber.NewCoreWebsocketConnection(sub.Endpoint)
	if err != nil {
		return nil, err
	}

	return &manager{
		jobid: sub.Job,
		conn:  conn,
	}, nil
}

func (m manager) Stop() {
	m.conn.Stop()
}
