package services

import (
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/smartcontractkit/external-initiator/subscriber"
)

type Starter func(sub store.Subscription, ch chan subscriber.Event, js *store.JobSpec) (Service, error)

type Service interface {
	Stop()
}
