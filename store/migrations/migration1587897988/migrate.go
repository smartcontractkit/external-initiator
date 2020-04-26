package migration1587897988

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576509489"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
)

type OntSubscription struct {
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
	Tezos        migration1576509489.TezosSubscription
	Substrate    migration1576783801.SubstrateSubscription
	Ontology     OntSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&OntSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate OntSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("ont_subscriptions").Error
}
