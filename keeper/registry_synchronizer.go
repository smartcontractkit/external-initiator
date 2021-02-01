package keeper

import (
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/smartcontractkit/external-initiator/store"
)

type RegistrySynchronizer interface {
	Start() error
	Stop()
}

type registrySynchronizer struct {
	endpoint  string
	ethClient *ethclient.Client
	chDone    chan struct{}
}

var _ RegistrySynchronizer = registrySynchronizer{}

func NewRegistrySynchronizer(config store.RuntimeConfig) RegistrySynchronizer {
	return registrySynchronizer{
		endpoint: config.KeeperEthEndpoint,
	}
}

func (rs registrySynchronizer) Start() error {
	ethClient, err := ethclient.Dial(rs.endpoint)
	if err != nil {
		return err
	}
	rs.ethClient = ethClient

	if !strings.HasPrefix(rs.endpoint, "ws") && !strings.HasPrefix(rs.endpoint, "http") {
		return fmt.Errorf("unknown endpoint protocol: %+v", rs.endpoint)
	}

	go rs.run()

	return nil
}

func (rs registrySynchronizer) Stop() {
	close(rs.chDone)
}

func (registrySynchronizer) run() {

}
