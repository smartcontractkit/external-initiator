package keeper

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/require"
)

var registryAddress = common.HexToAddress("0x0000000000000000000000000000000000000123")
var fromAddress = common.HexToAddress("0x0000000000000000000000000000000000000ABC")
var checkGasLimit = int64(10_000)
var cooldown = uint64(3)

func setupRegistrationManager(t *testing.T) (*store.Client, RegistrationManager, func()) {
	db, cleanup := store.SetupTestDB(t)
	rm := NewRegistrationManager(db.DB(), cooldown)
	return db, rm, cleanup
}

func newRegistrationWithUpkeepID(upkeepID int64) upkeepRegistration {
	return upkeepRegistration{
		UpkeepID:      upkeepID,
		Address:       registryAddress,
		From:          fromAddress,
		CheckGasLimit: checkGasLimit,
	}
}

// func TestRegistrationManager_PerformFullSync(t *testing.T) {
// 	db, rm, cleanup := setupRegistrationManager(t)
// 	defer cleanup()

// 	rm.PerformFullSync()
// 	// TODO - add client mocks

// 	assertRegistrationCount(t, db, 3)
// }

func TestRegistrationManager_Upsert(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	// create registration
	newRegistration := newRegistrationWithUpkeepID(0)
	err := rm.Upsert(newRegistration)
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
	var existingRegistration upkeepRegistration
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, checkGasLimit, existingRegistration.CheckGasLimit)
	require.Equal(t, int64(0), existingRegistration.LastRunBlockHeight)

	// update registration
	updatedRegistration := upkeepRegistration{
		UpkeepID:           0,
		Address:            registryAddress,
		From:               fromAddress,
		CheckGasLimit:      20_000,
		LastRunBlockHeight: 100,
	}
	err = rm.Upsert(updatedRegistration)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 1)
	err = db.DB().First(&existingRegistration).Error
	require.NoError(t, err)
	require.Equal(t, int64(20_000), existingRegistration.CheckGasLimit)
	require.Equal(t, int64(100), existingRegistration.LastRunBlockHeight)
}

func TestRegistrationManager_Delete(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	// create registration
	registration := newRegistrationWithUpkeepID(0)
	err := db.DB().Create(&registration).Error
	require.NoError(t, err)
	assertRegistrationCount(t, db, 1)

	// delete
	err = rm.Delete(registration.Address, 0)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)

	// delete again
	err = rm.Delete(registration.Address, 0)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)

	// delete a non-existent registration
	err = rm.Delete(registration.Address, 1234)
	require.NoError(t, err)
	assertRegistrationCount(t, db, 0)
}

func TestRegistrationManager_BatchDelete(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	registrations := [3]upkeepRegistration{
		newRegistrationWithUpkeepID(0),
		newRegistrationWithUpkeepID(1),
		newRegistrationWithUpkeepID(2),
	}

	for _, reg := range registrations {
		err := db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 3)

	err := rm.BatchDelete(registryAddress, []int64{0, 2})
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
}

func TestRegistrationManager_Active(t *testing.T) {
	db, rm, cleanup := setupRegistrationManager(t)
	defer cleanup()

	registrations := [3]upkeepRegistration{
		{ // valid
			UpkeepID:           0,
			Address:            registryAddress,
			From:               fromAddress,
			LastRunBlockHeight: 0, // 0 means never
			CheckGasLimit:      checkGasLimit,
		}, { // valid
			UpkeepID:           1,
			Address:            registryAddress,
			From:               fromAddress,
			LastRunBlockHeight: 6,
			CheckGasLimit:      checkGasLimit,
		}, { // too recent
			UpkeepID:           2,
			Address:            registryAddress,
			From:               fromAddress,
			LastRunBlockHeight: 7,
			CheckGasLimit:      checkGasLimit,
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
