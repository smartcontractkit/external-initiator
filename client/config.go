package client

import (
	"github.com/spf13/viper"
)

// Config contains the startup configuration parameters.
type Config struct {
	// The URL of the Chainlink Core service
	Chainlink string
	// ChainlinkAccessKey is the access key to identity the node to Chainlink
	ChainlinkAccessKey string
	// ChainlinkSecret is the secret to authenticate the node to Chainlink
	ChainlinkSecret string
	// DatabaseURL Configures the URL for chainlink to connect to. This must be
	// a properly formatted URL, with a valid scheme (postgres://).
	DatabaseURL string
}

// newConfigFromViper returns a Config based on the values supplied by viper.
func newConfigFromViper(v *viper.Viper) Config {
	return Config{
		Chainlink:          v.GetString("chainlink"),
		ChainlinkAccessKey: v.GetString("claccesskey"),
		ChainlinkSecret:    v.GetString("clsecret"),
		DatabaseURL:        v.GetString("databaseurl"),
	}
}
