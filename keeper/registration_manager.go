package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/jinzhu/gorm"
)

type RegistrationManager interface {
	Registries() ([]keeperRegistry, error)
	Upsert(upkeepRegistration) error
	SetRanAt(upkeepRegistration, uint64) error
	Delete(common.Address, uint64) error
	BatchDelete(int32, []uint64) error
	BatchCreate([]upkeepRegistration) error
	Active(chainHeight uint64) ([]upkeepRegistration, error)
}

func NewRegistrationManager(dbClient *gorm.DB, coolDown uint64) RegistrationManager {
	return registrationManager{
		dbClient: dbClient,
		coolDown: coolDown,
	}
}

type registrationManager struct {
	dbClient *gorm.DB
	coolDown uint64
}

// upkeepRegistration conforms to RegistrationManager interface
var _ RegistrationManager = registrationManager{}

func (rm registrationManager) Registries() (registries []keeperRegistry, _ error) {
	err := rm.dbClient.Find(&registries).Error
	return registries, err
}

func (rm registrationManager) Upsert(registration upkeepRegistration) error {
	return rm.dbClient.
		Set(
			"gorm:insert_option",
			`ON CONFLICT (registry_id, upkeep_id)
			DO UPDATE SET
				execute_gas = excluded.execute_gas,
				check_data = excluded.check_data
			`,
		).
		Create(&registration).
		Error
}

func (rm registrationManager) SetRanAt(registration upkeepRegistration, chainHeight uint64) error {
	registration.LastRunBlockHeight = chainHeight
	return rm.dbClient.Save(&registration).Error
}

func (rm registrationManager) Delete(address common.Address, upkeepID uint64) error {
	var registry keeperRegistry
	err := rm.dbClient.Where("address = ?", address).First(&registry).Error
	if err != nil {
		return err
	}

	return rm.dbClient.
		Table("upkeep_registrations").
		Where("registry_id = ? AND upkeep_id = ?", registry.ID, upkeepID).
		Delete("*").
		Error
}

func (rm registrationManager) BatchDelete(registryID int32, upkeedIDs []uint64) error {
	return rm.dbClient.
		Where("registry_id = ? AND upkeep_id IN (?)", registryID, upkeedIDs).
		Delete(upkeepRegistration{}).
		Error
}

func (rm registrationManager) BatchCreate(registrations []upkeepRegistration) error {
	return rm.dbClient.
		Create(&registrations).
		Error
}

func (rm registrationManager) Active(chainHeight uint64) (result []upkeepRegistration, _ error) {
	err := rm.dbClient.
		Where("last_run_block_height < ?", rm.runnableHeight(chainHeight)).
		Find(&result).
		Error

	return result, err
}

func (rm registrationManager) runnableHeight(chainHeight uint64) uint64 {
	if chainHeight < rm.coolDown {
		return 0
	}
	return chainHeight - rm.coolDown
}
