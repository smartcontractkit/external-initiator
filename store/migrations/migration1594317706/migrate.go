package migration1594317706

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576509489"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1587897988"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1592829052"
)

type NEARSubscription struct {
	gorm.Model
	SubscriptionId uint
	AccountIds     string
}

// TableName will set an explicit NEARSubscription table name, so table name isn't n_e_a_r_[...].
func (NEARSubscription) TableName() string {
	return "near_subscriptions"
}

type Subscription struct {
	gorm.Model
	ReferenceId       string `gorm:"unique;not null"`
	Job               string
	EndpointName      string
	Ethereum          migration0.EthSubscription
	Tezos             migration1576509489.TezosSubscription
	Substrate         migration1576783801.SubstrateSubscription
	Ontology          migration1587897988.OntSubscription
	BinanceSmartChain migration1592829052.BinanceSmartChainSubscription
	NEAR              NEARSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&NEARSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate NEARSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("near_subscriptions").Error // TODO: is this table ID correct? Where does it come from?
}
