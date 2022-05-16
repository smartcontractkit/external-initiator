package store

import (
	"database/sql"
	"database/sql/driver"
	"errors"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/smartcontractkit/chainlink/core/store/orm"
	"github.com/smartcontractkit/external-initiator/eitest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Config struct {
	DatabaseURL string
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

// prepareTestDB prepares the database to run tests, functionality varies
// on the underlying database.
func prepareTestDB(t *testing.T, config *Config) func() {
	t.Helper()
	return createPostgresChildDB(t, config, config.DatabaseURL)
}

func TestSQLStringArray_Scan(t *testing.T) {
	type args struct {
		src interface{}
	}
	tests := []struct {
		name    string
		arr     SQLStringArray
		args    args
		wantErr bool
		result  []string
	}{
		{
			"splits comma delimited string",
			SQLStringArray{},
			args{"abc,123"},
			false,
			[]string{"abc", "123"},
		},
		{
			"fails on invalid list",
			SQLStringArray{},
			args{`a""b,c`},
			true,
			nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.arr.Scan(tt.args.src); (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr {
				for i := range tt.result {
					assert.Equal(t, tt.result[i], tt.arr[i])
				}
			}
		})
	}
}

func TestSQLStringArray_Value(t *testing.T) {
	tests := []struct {
		name    string
		arr     SQLStringArray
		want    driver.Value
		wantErr bool
	}{
		{
			"turns string slice into csv list",
			SQLStringArray{"abc", "123"},
			"abc,123\n",
			false,
		},
		{
			"empty input gives empty output",
			SQLStringArray{},
			"\n",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.arr.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Value() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_SaveSubscription(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	sub := Subscription{
		ReferenceId:  "abc",
		Job:          "test123",
		EndpointName: "", // Missing name
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	assert.Error(t, err)

	sub = Subscription{
		ReferenceId:  "abc",
		Job:          "test123",
		EndpointName: "non-existent",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	assert.Error(t, err)

	sub = Subscription{
		ReferenceId:  "abc",
		Job:          "test123",
		EndpointName: "test",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	assert.NoError(t, err)

	oldSub := sub
	sub = Subscription{}
	subs, err := db.LoadSubscriptions()
	require.NoError(t, err)
	assert.Equal(t, 1, len(subs))
	assert.Equal(t, oldSub.ReferenceId, subs[0].ReferenceId)
	assert.Equal(t, oldSub.EndpointName, subs[0].Endpoint.Name)

	err = db.DeleteSubscription(&sub)
	assert.NoError(t, err)

	subs, err = db.LoadSubscriptions()
	assert.NoError(t, err)
	for _, s := range subs {
		assert.NotEqual(t, sub.ID, s.ID)
		assert.NotEqual(t, sub.ReferenceId, s.ReferenceId)
	}
}

func TestClient_SaveEndpoint(t *testing.T) {
	type args struct {
		endpoint *Endpoint
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"stores endpoint", args{endpoint: &Endpoint{
			Url:        "http://localhost:8545/",
			Type:       "ethereum",
			RefreshInt: 5,
			Name:       "eth-main",
		}}, false},
		{"overwrites name", args{endpoint: &Endpoint{
			Url:        "ws://localhost:8546/",
			Type:       "not-ethereum",
			RefreshInt: 0,
			Name:       "eth-main",
		}}, false},
	}

	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err = db.SaveEndpoint(tt.args.endpoint); (err != nil) != tt.wantErr {
				t.Errorf("SaveEndpoint() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && err == nil {
				e, err := db.LoadEndpoint(tt.args.endpoint.Name)
				require.NoError(t, err)
				assert.Equal(t, tt.args.endpoint.Name, e.Name)
				assert.Equal(t, tt.args.endpoint.Url, e.Url)
				assert.Equal(t, tt.args.endpoint.Type, e.Type)
				assert.Equal(t, tt.args.endpoint.RefreshInt, e.RefreshInt)
			}
		})
	}
}

func TestClient_prepareSubscription(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	sub := Subscription{
		ReferenceId:  "prepareTestA",
		Job:          "prepareTestA",
		EndpointName: "test",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	require.NoError(t, err)

	freshSub := Subscription{
		Model:        sub.Model,
		ReferenceId:  sub.ReferenceId,
		Job:          sub.Job,
		EndpointName: sub.EndpointName,
	}
	prepared, err := db.prepareSubscription(&freshSub)
	assert.NoError(t, err)
	assert.Equal(t, sub.Ethereum.Addresses, prepared.Ethereum.Addresses)
	assert.Equal(t, sub.Ethereum.Topics, prepared.Ethereum.Topics)
	assert.Equal(t, sub.EndpointName, prepared.Endpoint.Name)
}

func TestClient_LoadSubscription(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	jobId := "someJobId123"

	sub := Subscription{
		ReferenceId:  "LoadSubscriptionTestA",
		Job:          jobId,
		EndpointName: "test",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	require.NoError(t, err)

	res, err := db.LoadSubscription(jobId)
	assert.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, sub.ReferenceId, res.ReferenceId)
	assert.Equal(t, sub.Job, res.Job)

	_, err = db.LoadSubscription("invalid")
	assert.Error(t, err)
}

func TestClient_DeleteEndpoint(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	// Save test subscription that will be deleted
	// by DeleteEndpoint() call
	sub := Subscription{
		ReferenceId:  "DeleteEndpointTestA",
		Job:          "DeleteEndpointTestA",
		EndpointName: "test",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	require.NoError(t, err)

	err = db.DeleteEndpoint("test")
	assert.NoError(t, err)

	// Error should be returned when trying to
	// load our endpoint
	_, err = db.LoadEndpoint("test")
	assert.Error(t, err)

	// Subscription should have been deleted
	_, err = db.LoadSubscription(sub.Job)
	assert.Error(t, err)
}

func TestClient_DeleteAllEndpointsExcept(t *testing.T) {
	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	cleanupDB := prepareTestDB(t, &config)
	defer cleanupDB()
	db, err := ConnectToDb(config.DatabaseURL)
	require.NoError(t, err)
	defer eitest.MustClose(db)

	sub := Subscription{
		ReferenceId:  "DeleteAllEndpointsExceptTestA",
		Job:          "DeleteAllEndpointsExceptTestA",
		EndpointName: "test",
		Ethereum: EthSubscription{
			Addresses: []string{"0x12345"},
			Topics:    []string{"0xabcde"},
		},
	}
	err = db.SaveSubscription(&sub)
	require.NoError(t, err)

	newEndpoint := Endpoint{
		Url:  "http://localhost:8545/",
		Type: "ethereum",
		Name: "test2",
	}
	err = db.SaveEndpoint(&newEndpoint)
	assert.NoError(t, err)

	err = db.DeleteAllEndpointsExcept([]string{sub.EndpointName})
	assert.NoError(t, err)

	// No error loading the "test" endpoint
	// and associated subscription
	_, err = db.LoadEndpoint(sub.EndpointName)
	assert.NoError(t, err)
	_, err = db.LoadSubscription(sub.Job)
	assert.NoError(t, err)

	// Fails loading our newly created/deleted endpoint
	_, err = db.LoadEndpoint(newEndpoint.Name)
	assert.Error(t, err)
}
