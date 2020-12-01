package migration1605288480

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type EthCallSubscription struct {
	gorm.Model
	SubscriptionId   uint `gorm:"index"`
	Address          string
	ABI              string
	ResponseKey      string
	MethodName       string
	FunctionSelector []byte
	ReturnType       string
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&EthCallSubscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("eth_qae_subscriptions").Error
}
