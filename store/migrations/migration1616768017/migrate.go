package migration1616768017

import (
	"fmt"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type SubstrateSubscription struct {
	gorm.Model
	SubscriptionId uint   `gorm:"unique;not null"`
	AccountIds     string `gorm:"not null"`
	AccountId      uint   `gorm:"not null"`
	FeedId         string `gorm:"not null"`
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&SubstrateSubscription{}).Error
	fmt.Println("aAAA")
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate SubstrateSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("substrate_subscriptions").Error
}
