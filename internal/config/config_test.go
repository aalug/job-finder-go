package config

import (
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestLoadConfig(t *testing.T) {
	// Define test environment variables
	const (
		DBDriver             = "test_db_driver"
		DBSource             = "test_db_source"
		ServerAddress        = "test_server_address"
		ElasticSearchAddress = "test_elasticsearch_address"
		TokenSymmetricKey    = "test_token_symmetric_key"
		AccessTokenDuration  = "1h"
	)

	// Set the environment variables for testing
	setEnvVariables(t, map[string]string{
		"DB_DRIVER":             DBDriver,
		"DB_SOURCE":             DBSource,
		"SERVER_ADDRESS":        ServerAddress,
		"ELASTICSEARCH_ADDRESS": ElasticSearchAddress,
		"TOKEN_SYMMETRIC_KEY":   TokenSymmetricKey,
		"ACCESS_TOKEN_DURATION": AccessTokenDuration,
	})

	// Load the config
	config, err := LoadConfig("../../.")
	require.NoError(t, err)

	// require that the loaded configuration matches the environment variables
	require.Equal(t, DBDriver, config.DBDriver)
	require.Equal(t, DBSource, config.DBSource)
	require.Equal(t, ServerAddress, config.ServerAddress)
	require.Equal(t, ElasticSearchAddress, config.ElasticSearchAddress)
	require.Equal(t, TokenSymmetricKey, config.TokenSymmetricKey)

	expectedAccessTokenDuration, _ := time.ParseDuration(AccessTokenDuration)
	require.Equal(t, expectedAccessTokenDuration, config.AccessTokenDuration)

	// Reset the environment variables after the test
	resetEnvVariables()
}

// Helper function to set environment variables for testing
func setEnvVariables(t *testing.T, envVars map[string]string) {
	for key, value := range envVars {
		err := viper.BindEnv(key)
		require.NoError(t, err)
		viper.Set(key, value)
	}
}

// Helper function to reset environment variables after testing
func resetEnvVariables() {
	viper.Reset()
}
