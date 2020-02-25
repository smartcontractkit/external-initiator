package migration1582671289

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
)

func Migrate(tx *gorm.DB) error {
	err := tx.Model(&migration1576783801.Subscription{}).AddUniqueIndex("idx_job_id", "job").Error

	if err != nil {
		return errors.Wrap(err, "failed to add unique index to subscription job id")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.Model(&migration1576783801.Subscription{}).RemoveIndex("idx_job_id").Error
}
