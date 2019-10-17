package store

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/smartcontractkit/external-initiator/subscriber"
	"net/url"
)

type Client struct {
	db *badger.DB
}

func ConnectToDb(paths ...string) (Client, error) {
	path := "./db"
	if len(paths) > 0 {
		path = paths[0]
	}

	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return Client{}, err
	}

	return Client{db}, nil
}

func (client Client) Close() {
	client.db.Close()
}

func (client Client) loadPrefix(prefix []byte) ([][]byte, error) {
	var items [][]byte
	err := client.db.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			item := it.Item()
			val, err := item.ValueCopy(nil)
			if err != nil {
				return err
			}
			items = append(items, val)
		}
		return nil
	})
	return items, err
}

func (client Client) LoadSubscriptions() ([]Subscription, error) {
	var subs []Subscription
	items, err := client.loadPrefix([]byte("subscription-"))
	if err != nil {
		return subs, err
	}

	for _, item := range items {
		var sub Subscription
		err := json.Unmarshal(item, &sub)
		if err != nil {
			fmt.Println(err)
			continue
		}
		subs = append(subs, sub)
	}

	return subs, nil
}

func (client Client) SaveSubscription(sub Subscription) error {
	val, err := json.Marshal(sub)
	if err != nil {
		return err
	}

	return client.db.Update(func(txn *badger.Txn) error {
		err := txn.Set([]byte(fmt.Sprintf("subscription-%s", sub.Id)), val)
		return err
	})
}

type Endpoint struct {
	Url        url.URL
	Type       subscriber.Type
	Blockchain string
}

type Subscription struct {
	Id         string
	Job        string
	Addresses  []string
	Topics     []string
	Endpoint   Endpoint
	RefreshInt int
}
