package store

import "time"

type RuntimeConfig struct {
	FMAdapterTimeout       time.Duration
	FMAdapterRetryAttempts uint
	FMAdapterRetryDelay    time.Duration
}
