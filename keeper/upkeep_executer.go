package keeper

import (
	"net/url"

	"github.com/smartcontractkit/external-initiator/subscriber"
)

type UpkeepExecuter interface {
	Start() error
	Stop()
}

func NewUpkeepExecuter() UpkeepExecuter {
	return upkeepExecuter{}
}

type upkeepExecuter struct {
	endpoint   url.URL
	connection subscriber.Type
}

// upkeepExecuter satisfies UpkeepExecuter
var _ UpkeepExecuter = upkeepExecuter{}

func (executer upkeepExecuter) Start() error {
	// TODO - RYAN
	return nil
}

func (executer upkeepExecuter) Stop() {}
