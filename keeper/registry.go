package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/pborman/uuid"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

type keeperRegistry struct {
	ID          int32          `gorm:"primary_key"`
	Address     common.Address `gorm:"default:null"`
	CheckGas    uint32         `gorm:"default:null"`
	From        common.Address `gorm:"default:null"`
	JobID       *models.ID     `gorm:"default:null"`
	ReferenceID string         `gorm:"default:null"`
}

func NewRegistry(address common.Address, from common.Address, jobID *models.ID) keeperRegistry {
	return keeperRegistry{
		Address:     address,
		From:        from,
		JobID:       jobID,
		ReferenceID: uuid.New(),
	}
}
