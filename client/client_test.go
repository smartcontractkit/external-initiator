package client

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func Test_validateParams(t *testing.T) {
	t.Run("fails on missing required fields", func(t *testing.T) {
		v := viper.New()
		v.Set("required", "")
		err := validateParams(v, nil, []string{"required", "required2"})
		assert.Error(t, err)
	})

	t.Run("success with required fields", func(t *testing.T) {
		v := viper.New()
		v.Set("required", "value")
		v.Set("required2", "value")
		err := validateParams(v, nil, []string{"required", "required2"})
		assert.NoError(t, err)
	})

	t.Run("fails with invalid endpoint config", func(t *testing.T) {
		args := []string{
			`{"url":"http://localhost","name":"test"}`,
			`{invalid}`,
		}
		err := validateParams(viper.New(), args, nil)
		assert.Error(t, err)
	})

	t.Run("succeeds with valid endpoint config", func(t *testing.T) {
		args := []string{
			`{"url":"http://localhost","name":"test"}`,
			`{"url":"http://localhost","name":"valid"}`,
		}
		err := validateParams(viper.New(), args, nil)
		assert.NoError(t, err)
	})
}
