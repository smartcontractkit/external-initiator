package store

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/store/orm"
	"github.com/smartcontractkit/external-initiator/eitest"
	"github.com/stretchr/testify/require"
)

type Config struct {
	DatabaseURL string
}

func SetupTestDB(t *testing.T) (*Client, func()) {
	t.Helper()
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}
	cleanupDB := prepareTestDB(t, &config)
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)

	return db, func() { cleanupDB(); eitest.MustClose(db) }
}

// DropAndCreateThrowawayTestDB takes a database URL and appends the postfix to create a new database
func DropAndCreateThrowawayTestDB(databaseURL string, postfix string) (string, error) {
	parsed, err := url.Parse(databaseURL)
	if err != nil {
		return "", err
	}

	if parsed.Path == "" {
		return "", errors.New("path missing from database URL")
	}

	dbname := fmt.Sprintf("%s_%s", parsed.Path[1:], postfix)
	if len(dbname) > 62 {
		return "", errors.New("dbname too long, max is 63 bytes. Try a shorter postfix")
	}
	// Cannot drop test database if we are connected to it, so we must connect
	// to a different one. template1 should be present on all postgres installations
	parsed.Path = "/template1"
	db, err := sql.Open(string(orm.DialectPostgres), parsed.String())
	if err != nil {
		return "", fmt.Errorf("unable to open postgres database for creating test db: %+v", err)
	}
	defer eitest.MustClose(db)

	_, err = db.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbname))
	if err != nil {
		return "", fmt.Errorf("unable to drop postgres migrations test database: %v", err)
	}
	// `CREATE DATABASE $1` does not seem to work w CREATE DATABASE
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return "", fmt.Errorf("unable to create postgres migrations test database: %v", err)
	}
	parsed.Path = fmt.Sprintf("/%s", dbname)
	return parsed.String(), nil
}

func createTestDB(t *testing.T, parsed *url.URL) string {
	require.True(t, len(parsed.Path) > 1)

	path, err := DropAndCreateThrowawayTestDB(parsed.String(), fmt.Sprint(time.Now().Unix()))
	if err != nil {
		t.Fatal(err)
	}

	db, err := sql.Open(sqlDialect, path)
	if err != nil {
		t.Fatalf("unable to open postgres database for creating test db: %+v", err)
	}
	defer eitest.MustClose(db)

	return path
}

func seedTestDB(config Config) error {
	db, err := ConnectToDb(config.DatabaseURL)
	if err != nil {
		return err
	}
	defer eitest.MustClose(db)

	return db.db.Create(&Endpoint{Name: "test", Type: "ethereum", Url: "ws://localhost:8546/"}).Error
}

func createPostgresChildDB(t *testing.T, config *Config, originalURL string) func() {
	parsed, err := url.Parse(originalURL)
	if err != nil {
		t.Fatalf("unable to extract database from %v: %v", originalURL, err)
	}

	testdb := createTestDB(t, parsed)
	config.DatabaseURL = testdb

	if err = seedTestDB(*config); err != nil {
		t.Fatal(err)
	}

	return func() {
		config.DatabaseURL = testdb
	}
}

// PrepareTestDB prepares the database to run tests, functionality varies
// on the underlying database.
func prepareTestDB(t *testing.T, config *Config) func() {
	t.Helper()
	return createPostgresChildDB(t, config, config.DatabaseURL)
}
