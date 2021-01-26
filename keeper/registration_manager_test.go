package keeper

import (
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/smartcontractkit/chainlink/core/utils"
	"github.com/smartcontractkit/external-initiator/store"
	"github.com/stretchr/testify/require"
)

var address = common.HexToAddress("0x0000000000000000000000000000000000000123")
var checkGasLimit = int64(10_000)

func TestRegistrationManager_PerformFullSync(t *testing.T) {
	db, cleanup := store.SetupTestDB(t)
	defer cleanup()

	assertRegistrationCount(t, db, 0)

	rm := NewRegistrationManager(db)
	rm.PerformFullSync()
	// TODO - add client mocks

	assertRegistrationCount(t, db, 3)
}

func TestRegistrationManager_UpsertRegistration(t *testing.T) {
	db, cleanup := store.SetupTestDB(t)
	defer cleanup()

	assertRegistrationCount(t, db, 0)

	newRegistration := upkeepRegistration{
		UpkeepID:      utils.NewBigI(0),
		Address:       address,
		CheckGasLimit: checkGasLimit,
	}

	rm := NewRegistrationManager(db)
	err := rm.UpsertRegistration(newRegistration)
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
}

func TestRegistrationManager_DeleteRegistration(t *testing.T) {
	db, cleanup := store.SetupTestDB(t)
	defer cleanup()

	assertRegistrationCount(t, db, 0)

	registration := upkeepRegistration{
		UpkeepID:      utils.NewBigI(0),
		Address:       address,
		CheckGasLimit: checkGasLimit,
	}

	err := db.DB().Create(&registration).Error
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)

	manager := NewRegistrationManager(db)
	err = manager.DeleteRegistration(registration.Address, utils.NewBigI(0))
	require.NoError(t, err)

	assertRegistrationCount(t, db, 0)
}

func TestRegistrationManager_DeleteRegistrations(t *testing.T) {
	db, cleanup := store.SetupTestDB(t)
	defer cleanup()

	assertRegistrationCount(t, db, 0)
	address := common.HexToAddress("0x0000000000000000000000000000000000000123")

	registrations := [3]upkeepRegistration{
		{
			UpkeepID:      utils.NewBigI(0),
			Address:       address,
			CheckGasLimit: checkGasLimit,
		}, {
			UpkeepID:      utils.NewBigI(1),
			Address:       address,
			CheckGasLimit: checkGasLimit,
		}, {
			UpkeepID:      utils.NewBigI(2),
			Address:       address,
			CheckGasLimit: checkGasLimit,
		},
	}

	for _, reg := range registrations {
		err := db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 3)

	manager := NewRegistrationManager(db)
	err := manager.DeleteRegistrations(address, []utils.Big{*utils.NewBigI(0), *utils.NewBigI(2)})
	require.NoError(t, err)

	assertRegistrationCount(t, db, 1)
}

func TestRegistrationManager_GetActiveRegistrations(t *testing.T) {
	db, cleanup := store.SetupTestDB(t)
	defer cleanup()

	assertRegistrationCount(t, db, 0)

	// TODO
	// currentBlock := 10
	// coolDown := 3
	address := common.HexToAddress("0x0000000000000000000000000000000000000123")

	// valid
	registration1 := upkeepRegistration{
		UpkeepID:           utils.NewBigI(0),
		Address:            address,
		LastRunBlockHeight: 0, // 0 means never
		CheckGasLimit:      checkGasLimit,
	}
	// upkeep too recent
	registration2 := upkeepRegistration{
		UpkeepID:           utils.NewBigI(1),
		Address:            address,
		LastRunBlockHeight: 7,
		CheckGasLimit:      checkGasLimit,
	}

	for _, reg := range []upkeepRegistration{registration1, registration2} {
		err := db.DB().Create(&reg).Error
		require.NoError(t, err)
	}

	assertRegistrationCount(t, db, 2)

	manager := NewRegistrationManager(db)
	activeRegistrations, err := manager.GetActiveRegistrations()
	require.NoError(t, err)
	require.Len(t, activeRegistrations, 1)
	require.Equal(t, *big.NewInt(1), activeRegistrations[0].UpkeepID)
}

func assertRegistrationCount(t *testing.T, db *store.Client, expected int) {
	var count int
	db.DB().Model(&upkeepRegistration{}).Count(&count)
	require.Equal(t, expected, count)
}
