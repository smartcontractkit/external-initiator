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

type registrySynchronizer struct {
	endpoint            string
	ethClient           *ethclient.Client
	registrationManager RegistrationManager

	chDone chan struct{}
}

var _ RegistrySynchronizer = registrySynchronizer{}

func NewRegistrySynchronizer(dbClient *gorm.DB, config store.RuntimeConfig) RegistrySynchronizer {
	return registrySynchronizer{
		endpoint:            config.KeeperEthEndpoint,
		registrationManager: NewRegistrationManager(dbClient, uint64(config.KeeperBlockCooldown)),
	}
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
	existing, err := rs.registrationManager.UpkeepIDsForRegistry(registry)
	if err != nil {
		logger.Error(err)
	}

	existingSet := make(map[int64]bool)
	for _, upkeepID := range existing {
		existingSet[upkeepID] = true
	}

	contract, err := keeper_registry_contract.NewKeeperRegistryContract(registry.Address, rs.ethClient)
	if err != nil {
		logger.Error(err)
		return
	}

	// TODO - RYAN - DELETE can be done without loading eqisting id into memory
	// just do a delete where in (...)

	count, err := contract.GetUpkeepCount(nil)
	if err != nil {
		logger.Error(err)
		return
	}

	cancelled, err := contract.GetCanceledUpkeepList(nil)
	if err != nil {
		logger.Error(err)
		return
	}

	cancelledSet := make(map[int64]bool)
	for _, upkeepID := range cancelled {
		cancelledSet[upkeepID.Int64()] = true
	}

	var needToDelete []int64
	for _, upkeepID := range existing {
		if cancelledSet[upkeepID] {
			needToDelete = append(needToDelete, upkeepID)
		}
	}
	rs.registrationManager.BatchDelete(registry.ID, needToDelete)

	var needToUpsert []int64
	for upkeepID := int64(0); upkeepID < count.Int64(); upkeepID++ {
		if !cancelledSet[upkeepID] {
			needToUpsert = append(needToUpsert, upkeepID)
		}
	}

	for _, upkeepID := range needToUpsert {
		upkeepConfig, err := contract.GetUpkeep(nil, big.NewInt(upkeepID))
		if err != nil {
			logger.Error(err)
			continue
		}
		newUpkeep := upkeepRegistration{
			CheckGasLimit: uint64(upkeepConfig.ExecuteGas), // TODO - change to uint32?
			RegistryID:    uint32(registry.ID),
			UpkeepID:      upkeepID,
		}
		// TODO - RYAN - n+1 - parallelize
		rs.registrationManager.Upsert(newUpkeep)
	}
}
