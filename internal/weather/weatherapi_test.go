package weather_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"weatherApi/internal/weather"

	"github.com/stretchr/testify/require"
)

func TestGetCurrentWeather_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/current.json", r.URL.Path)
		require.Contains(t, r.URL.Query().Get("key"), "fake-api-key")
		require.Contains(t, r.URL.Query().Get("q"), "Kyiv")

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
	"current": {
		"temp_с": 21.5,
		"humidity": 55,
		"condition": {
			"text": "Clear"
		}
	}
}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	provider := weather.NewWeatherAPIProvider("fake-api-key", mockServer.URL)
	result, err := provider.GetWeather(context.Background(), "Kyiv")

	require.NoError(t, err)
	require.NotNil(t, result)
	require.Equal(t, 21.5, result.Temperature)
	require.Equal(t, 55, result.Humidity)
	require.Equal(t, "Clear", result.Description)
}

func TestCityExists_True(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
			"current": {
				"temp_с": 18.0,
				"humidity": 50,
				"condition": {
					"text": "Sunny"
				}
			}
		}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	provider := weather.NewWeatherAPIProvider("fake-api-key", mockServer.URL)
	exists, err := provider.CityIsValid(context.Background(), "Kyiv")

	require.NoError(t, err)
	require.True(t, exists)
}

func TestCityExists_False(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
			"error": {
				"message": "No matching location found."
			}
		}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	provider := weather.NewWeatherAPIProvider("fake-api-key", mockServer.URL)
	exists, err := provider.CityIsValid(context.Background(), "UnknownCity")

	require.ErrorIs(t, err, weather.ErrCityNotFound)
	require.False(t, exists)
}
