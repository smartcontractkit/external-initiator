package store

import "time"

type RuntimeConfig struct {
	KeeperBlockCooldown int64
	FMAdapterTimeout    time.Duration
}
