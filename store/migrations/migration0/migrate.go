package migration0

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
)

type Endpoint struct {
	gorm.Model
	Url        string
	Type       string
	RefreshInt int
	Name       string `gorm:"unique;not null"`
}

type Subscription struct {
	gorm.Model
	ReferenceId  string
	Job          string
	Addresses    string
	Topics       string
	EndpointName string
}

// Migrate runs the initial migration
func Migrate(tx *gorm.DB) error {
	err := tx.AutoMigrate(&Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&Endpoint{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Endpoint")
	}

	return nil
}
