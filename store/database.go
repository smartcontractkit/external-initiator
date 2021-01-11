// Package store encapsulates all database interaction.
package store

import (
	"bytes"
	"database/sql/driver"
	"encoding/csv"
	"fmt"
	"io"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/pkg/errors"
	"github.com/smartcontractkit/chainlink/core/logger"
	"github.com/smartcontractkit/external-initiator/store/migrations"
)

const sqlDialect = "postgres"

// SQLStringArray is a string array stored in the database as comma separated values.
type SQLStringArray []string

// Scan implements the sql Scanner interface.
func (arr *SQLStringArray) Scan(src interface{}) error {
	if src == nil {
		*arr = nil
	}
	v, err := driver.String.ConvertValue(src)
	if err != nil {
		return errors.New("failed to scan StringArray")
	}
	str, ok := v.(string)
	if !ok {
		return nil
	}

	buf := bytes.NewBufferString(str)
	r := csv.NewReader(buf)
	ret, err := r.Read()
	if err != nil && err != io.EOF {
		return errors.Wrap(err, "badly formatted csv string array")
	}
	*arr = ret
	return nil
}

// Value implements the driver Valuer interface.
func (arr SQLStringArray) Value() (driver.Value, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	err := w.Write(arr)
	if err != nil && err != io.EOF {
		return nil, errors.Wrap(err, "csv encoding of string array")
	}
	w.Flush()
	return buf.String(), nil
}

// SQLBytes is a byte slice stored in the database as a string.
type SQLBytes []byte

// Scan implements the sql Scanner interface.
func (bytes *SQLBytes) Scan(src interface{}) error {
	if src == nil {
		*bytes = nil
	}

	str, ok := src.(string)
	if !ok {
		return errors.New("failed to scan string")
	}

	*bytes = []byte(str)
	return nil
}

// Value implements the driver Valuer interface.
func (bytes SQLBytes) Value() (driver.Value, error) {
	return string(bytes), nil
}

// Client holds a connection to the database.
type Client struct {
	db *gorm.DB
}

// ConnectToDB attempts to connect to the database URI provided,
// and returns a new Client instance if successful.
func ConnectToDb(uri string) (*Client, error) {
	db, err := gorm.Open(sqlDialect, uri)
	if err != nil {
		return nil, fmt.Errorf("unable to open %s for gorm DB: %+v", uri, err)
	}
	if err = migrations.Migrate(db); err != nil {
		return nil, errors.Wrap(err, "newDBStore#Migrate")
	}
	store := &Client{
		db: db.Set("gorm:auto_preload", true),
	}
	return store, nil
}

// Close will close the connection to the database.
func (client Client) Close() error {
	return client.db.Close()
}

func (client Client) prepareSubscription(rawSub *Subscription) (*Subscription, error) {
	endpoint, err := client.LoadEndpoint(rawSub.EndpointName)
	if err != nil {
		return nil, err
	}

	sub := Subscription{
		Model:        rawSub.Model,
		ReferenceId:  rawSub.ReferenceId,
		Job:          rawSub.Job,
		EndpointName: rawSub.EndpointName,
		Endpoint:     endpoint,
	}

	switch endpoint.Type {
	case "ethereum", "iotex":
		if err := client.db.Model(&sub).Related(&sub.Ethereum).Error; err != nil {
			return nil, err
		}
	case "tezos":
		if err := client.db.Model(&sub).Related(&sub.Tezos).Error; err != nil {
			return nil, err
		}
	case "substrate":
		if err := client.db.Model(&sub).Related(&sub.Substrate).Error; err != nil {
			return nil, err
		}
	case "ontology":
		if err := client.db.Model(&sub).Related(&sub.Ontology).Error; err != nil {
			return nil, err
		}
	case "binance-smart-chain":
		if err := client.db.Model(&sub).Related(&sub.BinanceSmartChain).Error; err != nil {
			return nil, err
		}
	case "conflux":
		if err := client.db.Model(&sub).Related(&sub.Conflux).Error; err != nil {
			return nil, err
		}
	case "near":
		if err := client.db.Model(&sub).Related(&sub.NEAR).Error; err != nil {
			return nil, err
		}
	case "keeper":
		if err := client.db.Model(&sub).Related(&sub.Keeper).Error; err != nil {
			return nil, err
		}
	case "bsn-irita":
		if err := client.db.Model(&sub).Related(&sub.BSNIrita).Error; err != nil {
			return nil, err
		}
	}

	return &sub, nil
}

// LoadSubscriptions will find all subscriptions in the database,
// along with their associated endpoint and blockchain configuration,
// and return them in a slice.
func (client Client) LoadSubscriptions() ([]Subscription, error) {
	var sqlSubs []*Subscription
	if err := client.db.Find(&sqlSubs).Error; err != nil {
		return nil, err
	}

	var subs []Subscription
	for _, sqlSub := range sqlSubs {
		sub, err := client.prepareSubscription(sqlSub)
		if err != nil {
			logger.Error(err)
			continue
		}

		subs = append(subs, *sub)
	}

	return subs, nil
}

func (client Client) LoadSubscription(jobid string) (*Subscription, error) {
	var sub Subscription
	if err := client.db.Where("job = ?", jobid).First(&sub).Error; err != nil {
		return nil, err
	}

	return client.prepareSubscription(&sub)
}

// SaveSubscription will validate that the Endpoint exists,
// then store the Subscription in the database.
func (client Client) SaveSubscription(sub *Subscription) error {
	if len(sub.EndpointName) == 0 {
		sub.EndpointName = sub.Endpoint.Name
	}
	e, _ := client.LoadEndpoint(sub.EndpointName)
	if e.Name != sub.EndpointName {
		return fmt.Errorf("Unable to get endpoint %s", sub.EndpointName)
	}
	return client.db.Create(sub).Error
}

// DeleteSubscription will soft-delete the subscription provided.
func (client Client) DeleteSubscription(sub *Subscription) error {
	return client.db.Delete(sub).Error
}

// LoadEndpoint will return the endpoint in the database with
// the name provided.
func (client Client) LoadEndpoint(name string) (Endpoint, error) {
	var endpoint Endpoint
	err := client.db.Where(Endpoint{Name: name}).First(&endpoint).Error
	return endpoint, err
}

// RestoreEndpoint will restore any soft-deleted endpoint with
// the name provided.
func (client Client) RestoreEndpoint(name string) error {
	return client.db.Exec("UPDATE endpoints SET deleted_at = null WHERE name = ?", name).Error
}

// SaveEndpoint will store the endpoint in the database
// and overwrite any previous record with the same name.
func (client Client) SaveEndpoint(endpoint *Endpoint) error {
	err := client.db.Unscoped().Where(Endpoint{Name: endpoint.Name}).Assign(Endpoint{
		Url:        endpoint.Url,
		Type:       endpoint.Type,
		RefreshInt: endpoint.RefreshInt,
	}).FirstOrCreate(endpoint).Error
	if err != nil {
		return err
	}

	return client.RestoreEndpoint(endpoint.Name)
}

// DeleteEndpoint will soft-delete any endpoint record
// with the name provided, as well as any Subscriptions
// using this endpoint.
func (client Client) DeleteEndpoint(name string) error {
	err := client.db.Where(Endpoint{Name: name}).Delete(Endpoint{}).Error
	if err != nil {
		return err
	}

	// When deleting an endpoint, we should also delete
	// all subscriptions relying on this endpoint
	return client.db.Where(Subscription{EndpointName: name}).Delete(Subscription{}).Error
}

// DeleteAllEndpointsExcept will call DeleteEndpoint on all
// endpoints not provided.
func (client Client) DeleteAllEndpointsExcept(names []string) error {
	var endpoints []string
	err := client.db.Model(&Endpoint{}).Not("name", names).Pluck("name", &endpoints).Error
	if err != nil {
		return err
	}

	for _, name := range endpoints {
		err = client.DeleteEndpoint(name)
		if err != nil {
			return err
		}
	}

	return nil
}

type Endpoint struct {
	gorm.Model
	Url        string `json:"url"`
	Type       string `json:"type"`
	RefreshInt int    `json:"refreshInterval"`
	Name       string `json:"name"`
}

type Subscription struct {
	gorm.Model
	ReferenceId       string
	Job               string
	EndpointName      string
	Endpoint          Endpoint `gorm:"-"`
	Ethereum          EthSubscription
	Tezos             TezosSubscription
	Substrate         SubstrateSubscription
	Ontology          OntSubscription
	BinanceSmartChain BinanceSmartChainSubscription
	NEAR              NEARSubscription
	Conflux           CfxSubscription
	Keeper            KeeperSubscription
	BSNIrita          BSNIritaSubscription
}

type EthSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
	Topics         SQLStringArray
}

type TezosSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
}

type SubstrateSubscription struct {
	gorm.Model
	SubscriptionId uint
	AccountIds     SQLStringArray
}

type OntSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
}

type BinanceSmartChainSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
}

type NEARSubscription struct {
	gorm.Model
	SubscriptionId uint
	AccountIds     SQLStringArray
}

type CfxSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
	Topics         SQLStringArray
}

type KeeperSubscription struct {
	gorm.Model
	SubscriptionId uint
	Address        string
	UpkeepID       string
}

type BSNIritaSubscription struct {
	gorm.Model
	SubscriptionId uint
	Addresses      SQLStringArray
	ServiceName    string
}
