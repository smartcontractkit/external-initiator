package services

import "github.com/smartcontractkit/external-initiator/subscriber"

type RunlogManager struct {
	subscriber subscriber.ISubscriber
}

func NewRunlogManager(subscriber subscriber.ISubscriber) (*RunlogManager, error) {
	return nil, nil
}
