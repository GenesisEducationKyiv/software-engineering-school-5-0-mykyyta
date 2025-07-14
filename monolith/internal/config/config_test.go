package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestMustGet_PanicsWhenMissing(t *testing.T) {
	key := "MISSING_ENV_VAR"
	err := os.Unsetenv(key)
	require.NoError(t, err)

	assert.Panics(t, func() {
		_ = mustGet(key)
	})
}

func TestMustGet_ReturnsValue(t *testing.T) {
	key := "TEST_ENV_VAR"
	err := os.Setenv(key, "value123")
	require.NoError(t, err)

	defer func() {
		err := os.Unsetenv(key)
		require.NoError(t, err)
	}()

	val := mustGet(key)
	assert.Equal(t, "value123", val)
}

func TestGetEnv_ReturnsValueOrFallback(t *testing.T) {
	key := "OPTIONAL_ENV_VAR"

	err := os.Unsetenv(key)
	require.NoError(t, err)

	val := getEnv(key, "fallback")
	assert.Equal(t, "fallback", val)

	err = os.Setenv(key, "realvalue")
	require.NoError(t, err)

	defer func() {
		err := os.Unsetenv(key)
		require.NoError(t, err)
	}()

	val = getEnv(key, "fallback")
	assert.Equal(t, "realvalue", val)
}

func TestLoadConfig_DefaultsAndTrimming(t *testing.T) {
	// required vars for mustGet to succeed
	_ = os.Setenv("SENDGRID_API_KEY", "sendgrid-key")
	_ = os.Setenv("EMAIL_FROM", "me@example.com")
	_ = os.Setenv("WEATHER_API_KEY", "weather-key")
	_ = os.Setenv("TOMORROWIO_API_KEY", "tomorrowio-key")
	_ = os.Setenv("BASE_URL", "http://localhost:8080/") // note trailing slash

	cfg := LoadConfig()

	assert.Equal(t, "http://localhost:8080", cfg.BaseURL)
	assert.Equal(t, "8080", cfg.Port)
	assert.Equal(t, "debug", cfg.GinMode)
	assert.Equal(t, "postgres", cfg.DBType)
	assert.Equal(t, "sendgrid-key", cfg.SendGridKey)
	assert.Equal(t, "me@example.com", cfg.EmailFrom)
	assert.Equal(t, "weather-key", cfg.WeatherAPIKey)
}
