package migration0

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Endpoint struct {
	gorm.Model
	Url        string
	Type       string
	RefreshInt int
	Name       string `gorm:"unique;not null"`
}

type Subscription struct {
	gorm.Model
	ReferenceId  string `gorm:"unique;not null"`
	Job          string
	EndpointName string
	Ethereum     EthSubscription
}

type EthSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      string
	Topics         string
}

// Migrate runs the initial migration
func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&Endpoint{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Endpoint")
	}

	err = tx.AutoMigrate(&EthSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate EthSubscription")
	}

	return nil
}
