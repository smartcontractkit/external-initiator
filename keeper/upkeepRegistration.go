package keeper

type upkeepRegistration struct {
	ID                 int32  `gorm:"primary_key"`
	CheckGasLimit      uint64 `gorm:"default:null"`
	LastRunBlockHeight int64  `gorm:"not null;default:0"`
	RegistryID         uint32
	Registry           keeperRegistry
	UpkeepID           int64
}
