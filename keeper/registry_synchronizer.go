package keeper

import (
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/jinzhu/gorm"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/keeper/keeper_registry_contract"
	"github.com/smartcontractkit/external-initiator/store"
)

// TODO - RYAN - this should be an ENV VAR
// const registrySyncInterval = 5 * time.Minute
const registrySyncInterval = 10 * time.Second

type RegistrySynchronizer interface {
	Start() error
	Stop()
}

func NewRegistrySynchronizer(dbClient *gorm.DB, config store.RuntimeConfig) RegistrySynchronizer {
	return registrySynchronizer{
		endpoint:            config.KeeperEthEndpoint,
		registrationManager: NewRegistrationManager(dbClient, uint64(config.KeeperBlockCooldown)),
	}
}

type registrySynchronizer struct {
	endpoint            string
	ethClient           *ethclient.Client
	registrationManager RegistrationManager

	chDone chan struct{}
}

func (rs registrySynchronizer) Start() error {
	logger.Debug("starting registry synchronizer")
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

func (rs registrySynchronizer) run() {
	ticker := time.NewTicker(registrySyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rs.chDone:
			return
		case <-ticker.C:
			// TODO - RYAN - if sync takes too long? need a queue approach like in executer
			rs.performFullSync()
		}
	}
}

func (rs registrySynchronizer) performFullSync() {
	logger.Debug("performing full sync")

	registries, err := rs.registrationManager.Registries()
	if err != nil {
		logger.Error(err)
	}

	// TODO - parallellize this
	for _, registry := range registries {
		rs.syncRegistry(registry)
	}
}

func (rs registrySynchronizer) syncRegistry(registry keeperRegistry) {
	// WARN - this could get memory intensive depending on how many upkeeps there are

	contract, err := keeper_registry_contract.NewKeeperRegistryContract(registry.Address, rs.ethClient)
	if err != nil {
		logger.Error(err)
		return
	}

	count, err := contract.GetUpkeepCount(nil)
	if err != nil {
		logger.Error(err)
		return
	}

	cancelledBigs, err := contract.GetCanceledUpkeepList(nil)
	if err != nil {
		logger.Error(err)
		return
	}
	cancelled := make([]uint64, len(cancelledBigs))
	for idx, upkeepID := range cancelledBigs {
		cancelled[idx] = upkeepID.Uint64()
	}

	cancelledSet := make(map[uint64]bool)
	for _, upkeepID := range cancelled {
		cancelledSet[upkeepID] = true
	}

	err = rs.registrationManager.BatchDelete(registry.ID, cancelled)
	if err != nil {
		logger.Error(err)
	}

	var needToUpsert []uint64
	for upkeepID := uint64(0); upkeepID < count.Uint64(); upkeepID++ {
		if !cancelledSet[upkeepID] {
			needToUpsert = append(needToUpsert, upkeepID)
		}
	}

	for _, upkeepID := range needToUpsert {
		upkeepConfig, err := contract.GetUpkeep(nil, big.NewInt(int64(upkeepID)))
		if err != nil {
			logger.Error(err)
			continue
		}
		newUpkeep := upkeepRegistration{
			CheckData:  upkeepConfig.CheckData,
			ExecuteGas: upkeepConfig.ExecuteGas,
			RegistryID: uint32(registry.ID),
			UpkeepID:   upkeepID,
		}
		// TODO - RYAN - n+1 - parallelize
		err = rs.registrationManager.Upsert(newUpkeep)
		if err != nil {
			logger.Error(err)
		}
	}
}

func NewNoOpRegistrySynchronizer() RegistrySynchronizer {
	return noOpRegistrySynchronizer{}
}

type noOpRegistrySynchronizer struct{}

func (noOpRegistrySynchronizer) Start() error {
	return nil
}

func (noOpRegistrySynchronizer) Stop() {}
