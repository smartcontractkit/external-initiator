package store

import "time"

type RuntimeConfig struct {
	KeeperBlockCooldown    int64
	FMAdapterTimeout       time.Duration
	FMAdapterRetryAttempts uint
	FMAdapterRetryDelay    time.Duration
}
