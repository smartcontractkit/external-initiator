package migration1610281978

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576509489"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1587897988"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1592829052"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1594317706"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1599849837"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1608026935"
)

type BSNIritaSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      string
	ServiceName    string
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
	NEAR              migration1594317706.NEARSubscription
	Conflux           migration1599849837.CfxSubscription
	Keeper            migration1608026935.KeeperSubscription
	BSNIrita          BSNIritaSubscription
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&BSNIritaSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate BSNIritaSubscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("bsn_irita_subscriptions").Error
}
