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

type RegistrySynchronizer interface {
	Start() error
	Stop()
}

func NewRegistrySynchronizer(dbClient *gorm.DB, config store.RuntimeConfig) RegistrySynchronizer {
	return registrySynchronizer{
		endpoint:      config.KeeperEthEndpoint,
		registryStore: NewRegistryStore(dbClient, uint64(config.KeeperBlockCooldown)),
		interval:      config.KeeperRegistrySyncInterval,
	}
}

type registrySynchronizer struct {
	endpoint      string
	ethClient     *ethclient.Client
	interval      time.Duration
	registryStore RegistryStore

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
	ticker := time.NewTicker(rs.interval)
	defer ticker.Stop()

	for {
		select {
		case <-rs.chDone:
			return
		case <-ticker.C:
			// TODO - if sync takes too long? need a queue approach like in executer
			// https://www.pivotaltracker.com/story/show/176747117
			rs.performFullSync()
		}
	}
}

func (rs registrySynchronizer) performFullSync() {
	logger.Debug("performing full sync of keeper registries")

	registries, err := rs.registryStore.Registries()
	if err != nil {
		logger.Error(err)
	}

	// TODO - parallellize this
	// https://www.pivotaltracker.com/story/show/176747117
	for _, registry := range registries {
		rs.syncRegistry(registry)
	}
}

func (rs registrySynchronizer) syncRegistry(registry registry) {
	// WARN - this could get memory intensive depending on how many upkeeps there are

	contract, err := keeper_registry_contract.NewKeeperRegistryContract(registry.Address, rs.ethClient)
	if err != nil {
		logger.Error(err)
		return
	}

	// update registry config
	config, err := contract.GetConfig(nil)
	if err != nil {
		logger.Error(err)
		return
	}
	registry.CheckGas = config.CheckGasLimit
	registry.BlockCountPerTurn = uint32(config.BlockCountPerTurn.Uint64())
	err = rs.registryStore.UpdateRegistry(registry)
	if err != nil {
		logger.Error(err)
		return
	}

	// delete cancelled upkeeps
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
	err = rs.registryStore.BatchDelete(registry.ID, cancelled)
	if err != nil {
		logger.Error(err)
	}

	// add new upkeeps, update existing upkeeps
	count, err := contract.GetUpkeepCount(nil)
	if err != nil {
		logger.Error(err)
		return
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
		newUpkeep := registration{
			CheckData:  upkeepConfig.CheckData,
			ExecuteGas: upkeepConfig.ExecuteGas,
			RegistryID: uint32(registry.ID),
			UpkeepID:   upkeepID,
		}
		// TODO - parallelize
		// https://www.pivotaltracker.com/story/show/176747117
		err = rs.registryStore.Upsert(newUpkeep)
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
