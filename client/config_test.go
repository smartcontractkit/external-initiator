package client

import (
	"testing"

	"github.com/magiconair/properties/assert"
	"github.com/spf13/viper"
)

func Test_newConfigFromViper(t *testing.T) {
	t.Run("binds config variables", func(t *testing.T) {
		names := []string{"chainlinkurl", "ic_accesskey", "ic_secret", "databaseurl", "ci_accesskey", "ci_secret"}
		v := viper.New()
		for _, val := range names {
			v.Set(val, val)
		}

		conf := newConfigFromViper(v)
		assert.Equal(t, conf.ChainlinkURL, "chainlinkurl")
		assert.Equal(t, conf.InitiatorToChainlinkAccessKey, "ic_accesskey")
		assert.Equal(t, conf.InitiatorToChainlinkSecret, "ic_secret")
		assert.Equal(t, conf.DatabaseURL, "databaseurl")
		assert.Equal(t, conf.ChainlinkToInitiatorAccessKey, "ci_accesskey")
		assert.Equal(t, conf.ChainlinkToInitiatorSecret, "ci_secret")
	})
}
