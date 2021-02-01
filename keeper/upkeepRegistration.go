package keeper

type upkeepRegistration struct {
	ID                 int32  `gorm:"primary_key"`
	ExecuteGas         uint32 `gorm:"default:null"`
	LastRunBlockHeight uint64 `gorm:"not null;default:0"`
	RegistryID         uint32
	Registry           keeperRegistry
	UpkeepID           int64
}
