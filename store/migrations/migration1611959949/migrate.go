package migration1611959949

import (
	"github.com/jinzhu/gorm"
)

func Migrate(tx *gorm.DB) error {
	return tx.Exec(`
		ALTER TABLE keeper_subscriptions DROP COLUMN upkeep_id;
	`).Error
}

func Rollback(tx *gorm.DB) error {
	return tx.Exec(`
		ALTER TABLE keeper_subscriptions ADD COLUMN upkeep_id int;
	`).Error
}
