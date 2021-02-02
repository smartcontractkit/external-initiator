package keeper

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/google/uuid"
	"github.com/smartcontractkit/chainlink/core/store/models"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/require"
)

var registryAddress = common.HexToAddress("0x0000000000000000000000000000000000000123")
var fromAddress = common.HexToAddress("0x0000000000000000000000000000000000000ABC")
var checkDataStr = "ABC123"
var checkData = common.Hex2Bytes(checkDataStr)
var jobID = models.NewID()
var executeGas = uint32(10_000)
var checkGas = uint32(2_000_000)
var cooldown = uint64(3)

func setupRegistryStore(t *testing.T) (*store.Client, RegistryStore, func()) {
	db, cleanup := store.SetupTestDB(t)
	rm := NewRegistryStore(db.DB(), cooldown)
	return db, rm, cleanup
}

func newRegistry() registry {
	return registry{
		Address:     registryAddress,
		CheckGas:    checkGas,
		JobID:       jobID,
		From:        fromAddress,
		ReferenceID: uuid.New().String(),
	}
}

func newRegistration(reg registry, upkeepID uint64) registration {
	return registration{
		UpkeepID:   upkeepID,
		ExecuteGas: executeGas,
		Registry:   reg,
		CheckData:  checkData,
	}
}

func TestRegistryStore_Registries(t *testing.T) {
	db, rm, cleanup := setupRegistryStore(t)
	defer cleanup()

	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	reg2 := registry{
		Address:     common.HexToAddress("0x0000000000000000000000000000000000000456"),
		CheckGas:    checkGas,
		JobID:       models.NewID(),
		From:        fromAddress,
		ReferenceID: uuid.New().String(),
	}

	err = db.DB().Create(&reg2).Error
	require.NoError(t, err)

	existingRegistries, err := rm.Registries()
	require.NoError(t, err)
	require.Equal(t, 2, len(existingRegistries))
}

func TestRegistryStore_Upsert(t *testing.T) {
	db, rm, cleanup := setupRegistryStore(t)
	defer cleanup()

	// create registry
	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	// create registration
	newRegistration := newRegistration(reg, 0)
	fmt.Println("newRegistration.Registry.ID", newRegistration.Registry.ID)
	err = rm.Upsert(newRegistration)
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
	var existingRegistration registration
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, executeGas, existingRegistration.ExecuteGas)
	require.Equal(t, checkData, existingRegistration.CheckData)

	// update registration
	updatedRegistration := registration{
		Registry:   reg,
		UpkeepID:   0,
		ExecuteGas: 20_000,
		CheckData:  common.Hex2Bytes("8888"),
	}
	err = rm.Upsert(updatedRegistration)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 1)
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, uint32(20_000), existingRegistration.ExecuteGas)
	require.Equal(t, "8888", common.Bytes2Hex(existingRegistration.CheckData))
}

func TestRegistryStore_BatchDelete(t *testing.T) {
	db, rm, cleanup := setupRegistryStore(t)
	defer cleanup()

	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	registrations := [3]registration{
		newRegistration(reg, 0),
		newRegistration(reg, 1),
		newRegistration(reg, 2),
	}

	for _, reg := range registrations {
		err = db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 3)

	err = rm.BatchDelete(reg.ID, []uint64{0, 2})
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
}

func TestRegistryStore_Active(t *testing.T) {
	db, rm, cleanup := setupRegistryStore(t)
	defer cleanup()

	// create registry
	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	registrations := [3]registration{
		{ // valid
			UpkeepID:           0,
			LastRunBlockHeight: 0, // 0 means never
			ExecuteGas:         executeGas,
			Registry:           reg,
		}, { // valid
			UpkeepID:           1,
			LastRunBlockHeight: 6,
			ExecuteGas:         executeGas,
			Registry:           reg,
		}, { // too recent
			UpkeepID:           2,
			LastRunBlockHeight: 7,
			ExecuteGas:         executeGas,
			Registry:           reg,
		},
	}

	for _, reg := range registrations {
		err = db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 3)

	activeRegistrations, err := rm.Active(10)
	require.NoError(t, err)
	require.Len(t, activeRegistrations, 2)
	require.Equal(t, uint64(0), activeRegistrations[0].UpkeepID)
	require.Equal(t, uint64(1), activeRegistrations[1].UpkeepID)
}

func assertRegistrationCount(t *testing.T, db *store.Client, expected int) {
	var count int
	db.DB().Model(&registration{}).Count(&count)
	require.Equal(t, expected, count)
}
