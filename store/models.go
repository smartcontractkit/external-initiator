package store

import "time"

type RuntimeConfig struct {
	KeeperBlockCooldown        int64
	KeeperEthEndpoint          string
	KeeperRegistrySyncInterval time.Duration
}
