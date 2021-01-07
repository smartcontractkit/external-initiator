package client

import (
	"time"

	"github.com/spf13/viper"
)

// Config contains the startup configuration parameters.
type Config struct {
	// The URL of the ChainlinkURL Core Service
	ChainlinkURL string
	// InitiatorToChainlinkAccessKey is the access key to identity the node to ChainlinkURL
	InitiatorToChainlinkAccessKey string
	// InitiatorToChainlinkSecret is the secret to authenticate the node to ChainlinkURL
	InitiatorToChainlinkSecret string
	// DatabaseURL Configures the URL for chainlink to connect to. This must be
	// a properly formatted URL, with a valid scheme (postgres://).
	DatabaseURL string
	// The External Initiator access key, used for traffic flowing from Chainlink to this Service
	ChainlinkToInitiatorAccessKey string
	// The External Initiator secret, used for traffic flowing from Chainlink to this Service
	ChainlinkToInitiatorSecret string
	// ExpectsMock is true if the External Initiator should expect mock events from the blockchains
	ExpectsMock bool
	// ChainlinkTimeout sets the timeout for job run triggers to the Chainlink node
	ChainlinkTimeout time.Duration
	// ChainlinkRetryAttempts sets the maximum number of attempts that will be made for job run triggers
	ChainlinkRetryAttempts uint
	// ChainlinkRetryDelay sets the delay between attempts for job run triggers
	ChainlinkRetryDelay time.Duration
	// KeeperBlockCooldown sets a number of blocks to cool down before triggering a new run for a job.
	KeeperBlockCooldown int64
}

// newConfigFromViper returns a Config based on the values supplied by viper.
func newConfigFromViper(v *viper.Viper) Config {
	return Config{
		ChainlinkURL:                  v.GetString("chainlinkurl"),
		InitiatorToChainlinkAccessKey: v.GetString("ic_accesskey"),
		InitiatorToChainlinkSecret:    v.GetString("ic_secret"),
		DatabaseURL:                   v.GetString("databaseurl"),
		ChainlinkToInitiatorAccessKey: v.GetString("ci_accesskey"),
		ChainlinkToInitiatorSecret:    v.GetString("ci_secret"),
		ExpectsMock:                   v.GetBool("mock"),
		ChainlinkTimeout:              v.GetDuration("cl_timeout"),
		ChainlinkRetryAttempts:        v.GetUint("cl_retry_attempts"),
		ChainlinkRetryDelay:           v.GetDuration("cl_retry_delay"),
		KeeperBlockCooldown:           v.GetInt64("keeper_block_cooldown"),
	}
}
