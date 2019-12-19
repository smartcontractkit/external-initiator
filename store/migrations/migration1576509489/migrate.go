package migration1576509489

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
)

type TezosSubscription struct {
	gorm.Model
	SubscriptionId uint   `gorm:"unique;not null"`
	Addresses      string `gorm:"not null"`
}

type Subscription struct {
	gorm.Model
	ReferenceId  string `gorm:"unique;not null"`
	Job          string
	EndpointName string
	Ethereum     migration0.EthSubscription
	Tezos        TezosSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&TezosSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate TezosSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("tezos_subscriptions").Error
}
