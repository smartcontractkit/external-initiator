package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/badger"
	"log"
	"os"
	"testing"
)

var testDbPrefilled Client
var testDbEmpty Client
var subs []Subscription

func TestMain(m *testing.M) {
	db, err := ConnectToDb("/tmp/external-initiator-db-test")
	testDbPrefilled = db
	if err != nil {
		log.Fatal(err)
	}
	defer testDbPrefilled.Close()
	_ = testDbPrefilled.db.DropAll()

	testDbEmpty, err = ConnectToDb("/tmp/external-initiator-db-empty")
	if err != nil {
		log.Fatal(err)
	}
	defer testDbEmpty.Close()
	_ = testDbEmpty.db.DropAll()

	subs = []Subscription{
		{
			Id:  "abc",
			Job: "def",
		},
		{
			Id:  "123",
			Job: "456",
		},
		{
			Id:  "xyz",
			Job: "æøå",
		},
	}

	txn := testDbPrefilled.db.NewTransaction(true)
	for _, v := range subs {
		val, err := json.Marshal(v)
		if err != nil {
			log.Fatal(err)
		}

		if err := txn.Set([]byte(fmt.Sprintf("subscription-%s", v.Id)), val); err == badger.ErrTxnTooBig {
			_ = txn.Commit()
			txn = testDbPrefilled.db.NewTransaction(true)
			_ = txn.Set([]byte(fmt.Sprintf("subscription-%s", v.Id)), val)
		}
	}
	_ = txn.Commit()

	code := m.Run()
	os.Exit(code)
}

func TestClient_LoadSubscriptions(t *testing.T) {
	type fields struct {
		db *badger.DB
	}
	tests := []struct {
		name    string
		fields  fields
		results int
		wantErr bool
	}{
		{
			"no results",
			fields{db: testDbEmpty.db},
			0,
			false,
		},
		{
			"gives results",
			fields{db: testDbPrefilled.db},
			len(subs),
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				db: tt.fields.db,
			}
			got, err := client.LoadSubscriptions()
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadSubscriptions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.results != len(got) {
				t.Errorf("LoadSubscriptions() got %v results, want %v", len(got), tt.results)
			}
		})
	}
}

func TestClient_SaveSubscription(t *testing.T) {
	type fields struct {
		db *badger.DB
	}
	type args struct {
		sub Subscription
	}
	tests := []struct {
		name      string
		fields    fields
		args      args
		checkFunc func() error
	}{
		{
			"saves subscription",
			fields{db: testDbEmpty.db},
			args{Subscription{
				Id:  "uuid",
				Job: "randjobid",
			}},
			func() error {
				return testDbEmpty.db.View(func(txn *badger.Txn) error {
					item, err := txn.Get([]byte("subscription-uuid"))
					if err != nil {
						return err
					}

					return item.Value(func(val []byte) error {
						var sub Subscription
						err := json.Unmarshal(val, &sub)
						if err != nil {
							return err
						}

						if sub.Job != "randjobid" {
							return errors.New(fmt.Sprintf("Expected job id randjobid, got %s", sub.Job))
						}

						return nil
					})
				})
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				db: tt.fields.db,
			}
			if err := client.SaveSubscription(tt.args.sub); err != nil {
				t.Errorf("SaveSubscription() error = %v", err)
			}
			if err := tt.checkFunc(); err != nil {
				t.Errorf("SaveSubscription() checkFunc error = %v", err)
			}
		})
	}
}

func TestClient_loadPrefix(t *testing.T) {
	type fields struct {
		db *badger.DB
	}
	type args struct {
		prefix []byte
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    int
		wantErr bool
	}{
		{
			"gets all subscriptions",
			fields{db: testDbPrefilled.db},
			args{prefix: []byte(`subscription-`)},
			len(subs),
			false,
		},
		{
			"does not return error on invalid key",
			fields{db: testDbPrefilled.db},
			args{prefix: []byte(`doesnotexist-`)},
			0,
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := Client{
				db: tt.fields.db,
			}
			got, err := client.loadPrefix(tt.args.prefix)
			if (err != nil) != tt.wantErr {
				t.Errorf("loadPrefix() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want != len(got) {
				t.Errorf("LoadSubscriptions() got %v results, want %v", len(got), tt.want)
			}
		})
	}
}
