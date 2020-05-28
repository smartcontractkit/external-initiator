package migration1590000000

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576509489"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1587897988"
)

type CfxSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      string
	Topics         string
}

type Subscription struct {
	gorm.Model
	ReferenceId  string `gorm:"unique;not null"`
	Job          string
	EndpointName string
	Ethereum     migration0.EthSubscription
	Tezos        migration1576509489.TezosSubscription
	Substrate    migration1576783801.SubstrateSubscription
	Ontology     migration1587897988.OntSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&CfxSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate OntSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("ont_subscriptions").Error
}
