package migration1611603404

import (
	"github.com/jinzhu/gorm"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/external-initiator/store/migrations/migration1608026935"
)

func Migrate(tx *gorm.DB) error {
	return tx.Exec(`
		CREATE TABLE keeper_registries (
			id SERIAL PRIMARY KEY,
			reference_id uuid UNIQUE NOT NULL,
			address bytea UNIQUE NOT NULL,
			"from" bytea NOT NULL,
			check_gas int NOT NULL,
			block_count_per_turn int NOT NULL,
			job_id uuid UNIQUE NOT NULL
		);

		CREATE UNIQUE INDEX idx_keepers_unique_address ON keeper_registries(address);

		CREATE TABLE keeper_registrations (
			id SERIAL PRIMARY KEY,
			registry_id INT NOT NULL REFERENCES keeper_registries (id) ON DELETE CASCADE,
			execute_gas int NOT NULL,
			check_data bytea NOT NULL,
			last_run_block_height bigInt DEFAULT 0 NOT NULL,
			upkeep_id bigint NOT NULL
		);

		CREATE UNIQUE INDEX idx_keeper_registries_unique_jobs_per_registry ON keeper_registries(address, job_id);
		CREATE UNIQUE INDEX idx_keeper_registrations_unique_upkeep_ids_per_keeper ON keeper_registrations(registry_id, upkeep_id);

		DROP TABLE IF EXISTS keeper_subscriptions;
	`).Error
}

func Rollback(tx *gorm.DB) error {
	err := tx.Exec(`
		DROP TABLE IF EXISTS keeper_registries;
		DROP TABLE IF EXISTS keeper_registrations;
	`).Error

	if err != nil {
		return err
	}

	err = tx.AutoMigrate(&migration1608026935.Subscription{}).Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate Subscription")
	}

	err = tx.AutoMigrate(&migration1608026935.KeeperSubscription{}).AddForeignKey("subscription_id", "subscriptions(id)", "CASCADE", "CASCADE").Error
	if err != nil {
		return errors.Wrap(err, "failed to auto migrate EthQaeSubscription")
	}

	return nil
}
