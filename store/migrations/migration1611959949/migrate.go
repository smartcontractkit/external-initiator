package migration1611959949

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1608026935"
)

func Migrate(tx *gorm.DB) error {
	return tx.Exec(`
		DROP TABLE keeper_subscriptions;
	`).Error
}

func Rollback(tx *gorm.DB) error {
	err := tx.AutoMigrate(&migration1608026935.Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&migration1608026935.KeeperSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate EthQaeSubscription")
	}

	return nil
}
