package weatherapi

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"weatherApi/internal/weather"

	"github.com/stretchr/testify/require"
)

func TestGetCurrentWeather_Success(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/current.json", r.URL.Path)
		require.Equal(t, "fake-api-key", r.URL.Query().Get("key"))
		require.Equal(t, "Kyiv", r.URL.Query().Get("q"))

		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
			"current": {
				"temp_c": 21.5,
				"humidity": 55,
				"condition": {
					"text": "Clear"
				}
			}
		}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := mockServer.Client()
	provider := New("fake-api-key", client, mockServer.URL)

	result, err := provider.GetWeather(context.Background(), "Kyiv")

	require.NoError(t, err)
	require.Equal(t, 21.5, result.Temperature)
	require.Equal(t, 55, result.Humidity)
	require.Equal(t, "Clear", result.Description)
}

func TestCityExists_True(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
			"current": {
				"temp_c": 18.0,
				"humidity": 50,
				"condition": {
					"text": "Sunny"
				}
			}
		}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := mockServer.Client()
	provider := New("fake-api-key", client, mockServer.URL)

	exists, err := provider.CityIsValid(context.Background(), "Kyiv")

	require.NoError(t, err)
	require.True(t, exists)
}

func TestCityExists_False(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := fmt.Fprint(w, `{
			"error": {
				"code": 1006,
				"message": "No matching location found."
			}
		}`)
		require.NoError(t, err)
	}))
	defer mockServer.Close()

	client := mockServer.Client()
	provider := New("fake-api-key", client, mockServer.URL)

	exists, err := provider.CityIsValid(context.Background(), "UnknownCity")

	require.ErrorIs(t, err, weather.ErrCityNotFound)
	require.False(t, exists)
}

func TestGetCurrentWeather_Timeout(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Second)
	}))
	defer mockServer.Close()

	client := &http.Client{
		Timeout: 200 * time.Millisecond,
	}
	provider := New("fake-api-key", client, mockServer.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := provider.GetWeather(ctx, "Kyiv")

	require.Error(t, err)
	require.Contains(t, err.Error(), "timed out")
}
