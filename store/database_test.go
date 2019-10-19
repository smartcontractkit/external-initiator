package store

import (
	"database/sql"
	"fmt"
	"github.com/smartcontractkit/external-initiator/blockchain"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"testing"
	"time"
)

type Config struct {
	DatabaseURL string
}

func createTestDB(t *testing.T, parsed *url.URL) *url.URL {
	require.True(t, len(parsed.Path) > 1)
	dbname := fmt.Sprintf("%s_%d", parsed.Path[1:], time.Now().Unix())
	db, err := sql.Open(sqlDialect, parsed.String())
	if err != nil {
		t.Fatalf("unable to open postgres database for creating test db: %+v", err)
	}
	defer db.Close()

	//`CREATE DATABASE $1` does not seem to work w CREATE DATABASE
	_, err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		t.Fatalf("unable to create postgres test database: %+v", err)
	}

	newURL := *parsed
	newURL.Path = "/" + dbname
	return &newURL
}

func createPostgresChildDB(t *testing.T, config *Config, originalURL string) func() {
	parsed, err := url.Parse(originalURL)
	if err != nil {
		t.Fatalf("unable to extract database from %v: %v", originalURL, err)
	}

	testdb := createTestDB(t, parsed)
	config.DatabaseURL = testdb.String()

	return func() {
		reapPostgresChildDB(t, parsed, testdb)
		config.DatabaseURL = testdb.String()
	}
}

func reapPostgresChildDB(t *testing.T, parentURL, testURL *url.URL) {
	db, err := sql.Open(sqlDialect, parentURL.String())
	if err != nil {
		t.Fatalf("Unable to connect to parent CL db to clean up test database: %v", err)
	}
	defer db.Close()

	testdb := testURL.Path[1:]
	dbsSQL := "DROP DATABASE " + testdb
	_, err = db.Exec(dbsSQL)
	if err != nil {
		t.Fatalf("Unable to clean up previous test database: %v", err)
	}
}

// prepareTestDB prepares the database to run tests, functionality varies
// on the underlying database.
// Creates a second database, and returns a cleanup callback
// that drops said DB.
func prepareTestDB(t *testing.T, config *Config) func() {
	t.Helper()
	return createPostgresChildDB(t, config, config.DatabaseURL)
}

func TestClient_SaveSubscription(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	sub := Subscription{
		ReferenceId: "abc",
		Job:         "test123",
		Addresses:   []string{"0x12345"},
		Topics:      []string{"0xabcde"},
		Endpoint: Endpoint{
			Url:        "http://localhost/",
			Type:       int(subscriber.RPC),
			Blockchain: blockchain.ETH,
		},
		RefreshInt: 0,
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer db.Close()

	err = db.SaveSubscription(&sub)
	require.NoError(t, err)

	oldSub := sub
	sub = Subscription{}
	subs, err := db.LoadSubscriptions()
	require.NoError(t, err)
	assert.Equal(t, 1, len(subs))
	assert.Equal(t, oldSub.ReferenceId, subs[0].ReferenceId)
	assert.Equal(t, oldSub.Endpoint.Url, subs[0].Endpoint.Url)
}
