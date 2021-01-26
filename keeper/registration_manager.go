package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/external-initiator/store"
)

type RegistrationManager interface {
	PerformFullSync() error
	UpsertRegistration(upkeepRegistration) error
	DeleteRegistration(common.Address, *utils.Big) error
	DeleteRegistrations(common.Address, []utils.Big) error
	GetActiveRegistrations() ([]upkeepRegistration, error)
}

func NewRegistrationManager(dbClient *store.Client) RegistrationManager {
	return registrationManager{dbClient}
}

type registrationManager struct {
	dbClient *store.Client
}

type upkeepRegistration struct {
	UpkeepID           *utils.Big
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

func (rm registrationManager) UpsertRegistration(registration upkeepRegistration) error {
	return rm.dbClient.DB().
		Set(
			"gorm:insert_option",
			`ON CONFLICT (address, upkeep_id)
			DO UPDATE SET check_gas_limit = excluded.check_gas_limit, last_run_block_height = excluded.last_run_block_height`,
		).
		Create(&registration).
		Error
}

func (registrationManager) DeleteRegistration(common.Address, *utils.Big) error {
	// TODO
	return nil
}

func (registrationManager) DeleteRegistrations(common.Address, []utils.Big) error {
	// TODO
	return nil
}

func (registrationManager) GetActiveRegistrations() ([]upkeepRegistration, error) {
	// TODO
	return nil, nil
}
