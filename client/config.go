package client

import (
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
	}
}
