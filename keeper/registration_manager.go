package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/core/utils"
)

type RegistrationManager interface {
	PerformFullSync() error
	UpsertRegistration(upkeepRegistration) error
	DeleteRegistration(common.Address, *utils.Big) error
	DeleteRegistrations(common.Address, []utils.Big) error
	GetActiveRegistrations() ([]upkeepRegistration, error)
}

func NewRegistrationManager() RegistrationManager {
	return registrationManager{}
}

type registrationManager struct {
}

type upkeepRegistration struct {
	UpkeepID      *utils.Big
	Address       common.Address
	LastRun       int64 `gorm:"not null;default:0"`
	CheckGasLimit int64
}

// upkeepRegistration conforms to RegistrationManager interface
var _ RegistrationManager = registrationManager{}

func (registrationManager) PerformFullSync() error {
	// TODO
	return nil
}

func (registrationManager) UpsertRegistration(upkeepRegistration) error {
	// TODO
	return nil
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
