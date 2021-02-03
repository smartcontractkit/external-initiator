package keeper

import (
	"github.com/jinzhu/gorm"
)

type RegistryStore interface {
	Registries() ([]registry, error)
	UpdateRegistry(registry registry) error
	Upsert(registration) error
	UpdateRanAt(registration, uint64) error
	BatchDelete(registryID uint32, upkeedIDs []uint64) error
	Active(chainHeight uint64) ([]registration, error)
}

func NewRegistryStore(dbClient *gorm.DB, coolDown uint64) RegistryStore {
	return registryStore{
		dbClient: dbClient,
		coolDown: coolDown,
	}
}

type registryStore struct {
	dbClient *gorm.DB
	coolDown uint64
}

func (rm registryStore) Registries() (registries []registry, _ error) {
	err := rm.dbClient.Find(&registries).Error
	return registries, err
}

func (rm registryStore) UpdateRegistry(registry registry) error {
	return rm.dbClient.Save(&registry).Error
}

func (rm registryStore) Upsert(registration registration) error {
	return rm.dbClient.
		Set(
			"gorm:insert_option",
			`ON CONFLICT (registry_id, upkeep_id)
			DO UPDATE SET
				execute_gas = excluded.execute_gas,
				check_data = excluded.check_data
			`,
		).
		Set("gorm:save_associations", false).
		Create(&registration).
		Error
}

func (rm registryStore) UpdateRanAt(registration registration, chainHeight uint64) error {
	registration.LastRunBlockHeight = chainHeight
	return rm.dbClient.Save(&registration).Error
}

func (rm registryStore) BatchDelete(registryID uint32, upkeedIDs []uint64) error {
	return rm.dbClient.
		Where("registry_id = ? AND upkeep_id IN (?)", registryID, upkeedIDs).
		Delete(registration{}).
		Error
}

func (rm registryStore) Active(chainHeight uint64) (result []registration, _ error) {
	err := rm.dbClient.
		Where("last_run_block_height < ?", rm.runnableHeight(chainHeight)).
		Find(&result).
		Error

	return result, err
}

func (rm registryStore) runnableHeight(chainHeight uint64) uint64 {
	if chainHeight < rm.coolDown {
		return 0
	}
	return chainHeight - rm.coolDown
}
