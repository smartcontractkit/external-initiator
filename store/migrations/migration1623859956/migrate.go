package migration1623859956

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type TerraSubscription struct {
	gorm.Model
	SubscriptionId  uint `gorm:"unique;not null"`
	ContractAddress string
	AccountAddress  string
}

type Subscription struct {
	gorm.Model
	ReferenceId       string `gorm:"unique;not null"`
	Job               string
	EndpointName      string
	Terra             TerraSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&TerraSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TerraSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("terra_subscriptions").Error
}
