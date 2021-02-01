package migration1611603404

import (
	"github.com/jinzhu/gorm"
)

func Migrate(tx *gorm.DB) error {
	return tx.Exec(`
		CREATE TABLE keeper_registries (
			id SERIAL PRIMARY KEY,
			reference_id uuid UNIQUE NOT NULL,
			address bytea UNIQUE NOT NULL,
			"from" bytea NOT NULL,
			job_id uuid UNIQUE NOT NULL
		);

		CREATE UNIQUE INDEX idx_keepers_unique_address ON keeper_registries(address);

		CREATE TABLE upkeep_registrations (
			id SERIAL PRIMARY KEY,
			registry_id INT NOT NULL REFERENCES keeper_registries (id) ON DELETE CASCADE,
			execute_gas int NOT NULL,
			last_run_block_height bigInt DEFAULT 0 NOT NULL,
			upkeep_id bigint NOT NULL
		);

		CREATE UNIQUE INDEX idx_upkeep_registrations_unique_upkeep_ids_per_keeper ON upkeep_registrations(registry_id, upkeep_id);
	`).Error
}

func Rollback(tx *gorm.DB) error {
	return tx.Exec(`
		DROP TABLE IF EXISTS keeper_registries;
		DROP TABLE IF EXISTS upkeep_registrations;
	`).Error
}
