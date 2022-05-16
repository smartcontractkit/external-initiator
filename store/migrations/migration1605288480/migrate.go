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
<<<<<<< HEAD
	FunctionSelector []byte
=======
	FunctionSelector [4]byte
>>>>>>> c8a538596d61f226ae9c220962667ff5379119d6
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
<<<<<<< HEAD
	return tx.DropTable("eth_qae_subscriptions").Error
=======
	return tx.DropTable("eth_call_subscriptions").Error
>>>>>>> c8a538596d61f226ae9c220962667ff5379119d6
}
