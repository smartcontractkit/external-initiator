package migrations

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration0"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576509489"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1576783801"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1582671289"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1587897988"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1592829052"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1594317706"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1590000000"
	"gopkg.in/gormigrate.v1"
)

// Migrate iterates through available migrations, running and tracking
// migrations that have not been run.
func Migrate(db *gorm.DB) error {
	options := *gormigrate.DefaultOptions
	options.UseTransaction = true

	migrations := []*gormigrate.Migration{
		{
			ID:      "0",
			Migrate: migration0.Migrate,
		},
		{
			ID:       "1576509489",
			Migrate:  migration1576509489.Migrate,
			Rollback: migration1576509489.Rollback,
		},
		{
			ID:       "1576783801",
			Migrate:  migration1576783801.Migrate,
			Rollback: migration1576783801.Rollback,
		},
		{
			ID:       "1582671289",
			Migrate:  migration1582671289.Migrate,
			Rollback: migration1582671289.Rollback,
		},
		{
			ID:       "1587897988",
			Migrate:  migration1587897988.Migrate,
			Rollback: migration1587897988.Rollback,
		},
		{
			ID:       "1592829052",
			Migrate:  migration1592829052.Migrate,
			Rollback: migration1592829052.Rollback,
		},
		{
			ID:       "1594317706",
			Migrate:  migration1594317706.Migrate,
			Rollback: migration1594317706.Rollback,
		},
		{
			ID:       "1590000000",
			Migrate:  migration1590000000.Migrate,
			Rollback: migration1590000000.Rollback,
		},
	}

	m := gormigrate.New(db, &options, migrations)

	err := m.Migrate()
	if err != nil {
		return errors.Wrap(err, "error running migrations")
	}
	return nil
}
