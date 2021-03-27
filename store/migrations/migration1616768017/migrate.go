package migration1616768017

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type SubstrateSubscription struct {
	gorm.Model
	SubscriptionId uint   `gorm:"unique;not null"`
	AccountIds     string `gorm:"not null"`
	AccountId      string `gorm:"not null"`
	FeedId         uint32 `gorm:"not null"`
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&SubstrateSubscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate SubstrateSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("substrate_subscriptions").Error
}
