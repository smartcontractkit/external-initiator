package migration1616788127

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type TerraSubscription struct {
	gorm.Model
	ContractAddress string `gorm:"not null"`
	AccountAddress  string `gorm:"not null"`
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&TerraSubscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TerraSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("terra_subscriptions").Error
}
