package migration1611169747

import (
	"github.com/jinzhu/gorm"
)

func Migrate(tx *gorm.DB) error {
	return tx.Exec(`
		ALTER TABLE keeper_subscriptions ADD COLUMN "from" bytea NOT NULL;
	`).Error
}

func Rollback(tx *gorm.DB) error {
	return tx.Exec(`
		ALTER TABLE keeper_subscriptions DROP COLUMN IF EXISTS "from";
	`).Error
}
