package keeper

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/core/store/models"
)

type keeperRegistry struct {
	ID          int32          `gorm:"primary_key"`
	Address     common.Address `gorm:"default:null"`
	From        common.Address `gorm:"default:null"`
	JobID       *models.ID     `gorm:"default:null"`
	ReferenceID string         `gorm:"default:null"`
}
