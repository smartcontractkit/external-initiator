package migration1613356333

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type JobSpec struct {
	gorm.Model
	Job  string `gorm:"unique;not null"`
	Spec json.RawMessage
}

func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&JobSpec{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	return nil
}

func Rollback(tx *gorm.DB) error {
	return tx.DropTable("job_specs").Error
}
