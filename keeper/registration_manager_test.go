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
var jobID = models.NewID()
var checkGasLimit = uint64(10_000)
var cooldown = uint64(3)

func setupRegistrationManager(t *testing.T) (*store.Client, RegistrationManager, func()) {
	db, cleanup := store.SetupTestDB(t)
	rm := NewRegistrationManager(db.DB(), cooldown)
	return db, rm, cleanup
}

func newRegistry() keeperRegistry {
	return keeperRegistry{
		Address:     registryAddress,
		JobID:       jobID,
		From:        fromAddress,
		ReferenceID: uuid.New().String(),
	}
}

func newRegistration(reg keeperRegistry, upkeepID int64) upkeepRegistration {
	return upkeepRegistration{
		UpkeepID:      upkeepID,
		CheckGasLimit: checkGasLimit,
		Registry:      reg,
	}
}

func TestRegistrationManager_Registries(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	reg2 := keeperRegistry{
		Address:     common.HexToAddress("0x0000000000000000000000000000000000000456"),
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

func TestRegistrationManager_Upsert(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
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
	var existingRegistration upkeepRegistration
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, checkGasLimit, existingRegistration.CheckGasLimit)
	require.Equal(t, int64(0), existingRegistration.LastRunBlockHeight)

	// update registration
	updatedRegistration := upkeepRegistration{
		Registry:           reg,
		UpkeepID:           0,
		CheckGasLimit:      20_000,
		LastRunBlockHeight: 100,
	}
	err = rm.Upsert(updatedRegistration)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 1)
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, uint64(20_000), existingRegistration.CheckGasLimit)
	require.Equal(t, int64(100), existingRegistration.LastRunBlockHeight)
}

func TestRegistrationManager_Delete(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	// create registry
	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	// create registration
	registration := newRegistration(reg, 0)
	err = db.DB().Create(&registration).Error
	require.NoError(t, err)
	assertRegistrationCount(t, db, 1)

	// delete
	err = rm.Delete(registration.Registry.Address, 0)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)

	// delete again (we don't want it to error)
	err = rm.Delete(registration.Registry.Address, 0)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)

	// delete a non-existent registration
	err = rm.Delete(registration.Registry.Address, 1234)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)
}

func TestRegistrationManager_BatchDelete(t *testing.T) {
	// db, rm, cleanup := setupRegistrationManager(t)
	// defer cleanup()

	// registrations := [3]upkeepRegistration{
	// 	newRegistrationWithUpkeepID(0),
	// 	newRegistrationWithUpkeepID(1),
	// 	newRegistrationWithUpkeepID(2),
	// }

	// for _, reg := range registrations {
	// 	err := db.DB().Create(&reg).Error
	// 	require.NoError(t, err)
	// }

	// assertRegistrationCount(t, db, 3)

	// err := rm.BatchDelete(registryAddress, []int64{0, 2})
	// require.NoError(t, err)

	// assertRegistrationCount(t, db, 1)
}

func TestRegistrationManager_Active(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	// create registry
	reg := newRegistry()
	err := db.DB().Create(&reg).Error
	require.NoError(t, err)

	registrations := [3]upkeepRegistration{
		{ // valid
			UpkeepID:           0,
			LastRunBlockHeight: 0, // 0 means never
			CheckGasLimit:      checkGasLimit,
			Registry:           reg,
		}, { // valid
			UpkeepID:           1,
			LastRunBlockHeight: 6,
			CheckGasLimit:      checkGasLimit,
			Registry:           reg,
		}, { // too recent
			UpkeepID:           2,
			LastRunBlockHeight: 7,
			CheckGasLimit:      checkGasLimit,
			Registry:           reg,
		},
	}

	for _, reg := range registrations {
		err := db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 3)

	activeRegistrations, err := rm.Active(10)
	require.NoError(t, err)
	require.Len(t, activeRegistrations, 2)
	require.Equal(t, int64(0), activeRegistrations[0].UpkeepID)
	require.Equal(t, int64(1), activeRegistrations[1].UpkeepID)
}

func assertRegistrationCount(t *testing.T, db *store.Client, expected int) {
	var count int
	db.DB().Model(&upkeepRegistration{}).Count(&count)
	require.Equal(t, expected, count)
}
