package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/uuid"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

type registry struct {
	ID                uint32         `gorm:"primary_key"`
	Address           common.Address `gorm:"default:null"`
	BlockCountPerTurn uint32
	CheckGas          uint32
	From              common.Address `gorm:"default:null"`
	JobID             *models.ID     `gorm:"default:null"`
	ReferenceID       string         `gorm:"default:null"`
}

func NewRegistry(address common.Address, from common.Address, jobID *models.ID) registry {
	return registry{
		Address:     address,
		From:        from,
		JobID:       jobID,
		ReferenceID: uuid.New(),
	}
}

func (registry) TableName() string {
	return "keeper_registries"
}
