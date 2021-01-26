package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/external-initiator/store"
)

type RegistrationManager interface {
	PerformFullSync() error
	Upsert(upkeepRegistration) error
	Delete(common.Address, int64) error
	BatchDelete(common.Address, []int64) error
	Active(uint) ([]upkeepRegistration, error)
}

func NewRegistrationManager(dbClient *store.Client, coolDown uint) RegistrationManager {
	return registrationManager{
		dbClient: dbClient,
		coolDown: coolDown,
	}
}

type registrationManager struct {
	dbClient          *store.Client
	coolDown          uint
	latestBlockHeight uint
}

type upkeepRegistration struct {
	UpkeepID           int64
	Address            common.Address
	LastRunBlockHeight int64 `gorm:"not null;default:0"`
	CheckGasLimit      int64
}

// upkeepRegistration conforms to RegistrationManager interface
var _ RegistrationManager = registrationManager{}

func (registrationManager) PerformFullSync() error {
	// TODO
	return nil
}

func (rm registrationManager) Upsert(registration upkeepRegistration) error {
	return rm.dbClient.DB().
		Set(
			"gorm:insert_option",
			`ON CONFLICT (address, upkeep_id)
			DO UPDATE SET check_gas_limit = excluded.check_gas_limit, last_run_block_height = excluded.last_run_block_height`,
		).
		Create(&registration).
		Error
}

func (rm registrationManager) Delete(address common.Address, upkeepID int64) error {
	return rm.dbClient.DB().
		Where("address = ? AND upkeep_id = ?", address, upkeepID).
		Delete(upkeepRegistration{}).
		Error
}

func (rm registrationManager) BatchDelete(address common.Address, upkeedIDs []int64) error {
	return rm.dbClient.DB().
		Where("address = ? AND upkeep_id IN (?)", address, upkeedIDs).
		Delete(upkeepRegistration{}).
		Error
}

func (rm registrationManager) Active(chainHeight uint) (result []upkeepRegistration, _ error) {
	err := rm.dbClient.DB().
		Where("last_run_block_height < ?", rm.runnableHeight(chainHeight)).
		Find(&result).
		Error

	return result, err
}

func (rm registrationManager) runnableHeight(chainHeight uint) uint {
	if chainHeight < rm.coolDown {
		return 0
	}
	return chainHeight - rm.coolDown
}
