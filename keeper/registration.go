package keeper

type registration struct {
	ID                 int32 `gorm:"primary_key"`
	CheckData          []byte
	ExecuteGas         uint32 `gorm:"default:null"`
	LastRunBlockHeight uint64 `gorm:"not null;default:0"`
	RegistryID         uint32
	Registry           registry
	UpkeepID           uint64
}

func (registration) TableName() string {
	return "keeper_registrations"
}
