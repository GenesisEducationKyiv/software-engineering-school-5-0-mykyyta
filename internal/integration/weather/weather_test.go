//go:build integration

package weather_test

import (
	"context"
	"log"
	"os"
	"strings"
	"testing"

	"weatherApi/internal/weather"

	"github.com/stretchr/testify/require"
)

// --- Fake Providers ---

type successProvider struct{}

func (s *successProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{Temperature: 20.5, Humidity: 60, Description: "Sunny"}, nil
}

func (s *successProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}

type failProvider struct{}

func (f *failProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{}, context.DeadlineExceeded
}

func (f *failProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, context.DeadlineExceeded
}

type cityNotFoundProvider struct{}

func (c *cityNotFoundProvider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	return weather.Report{}, weather.ErrCityNotFound
}

func (c *cityNotFoundProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, weather.ErrCityNotFound
}

// --- Test Utils ---

func chainWith(handlers ...weather.Handler) weather.Handler {
	if len(handlers) == 0 {
		return nil
	}
	head := handlers[0]
	current := head
	for _, next := range handlers[1:] {
		current = current.SetNext(next)
	}
	return head
}

// --- Tests ---

func TestIntegration_ChainAndLogging(t *testing.T) {
	logFile := "test_weather.log"
	defer os.Remove(logFile)

	file, err := os.Create(logFile)
	require := require.New(t)
	require.NoError(err)

	defer file.Close()
	logger := log.New(file, "", log.LstdFlags)

	fp := &failProvider{}
	sp := &successProvider{}
	wrappedFail := weather.NewWrapper(fp, "FailProvider", logger)
	wrappedSuccess := weather.NewWrapper(sp, "SuccessProvider", logger)

	handler := chainWith(
		weather.NewBaseProvider(wrappedFail),
		weather.NewBaseProvider(wrappedSuccess),
	)

	ctx := context.Background()
	report, err := handler.GetWeather(ctx, "Kyiv")
	require.NoError(err)
	require.Equal("Sunny", report.Description)

	file.Sync()
	data, err := os.ReadFile(logFile)
	require.NoError(err)

	content := string(data)
	if !strings.Contains(content, "GetWeather") || !strings.Contains(content, "Kyiv") {
		t.Errorf("log does not contain expected entries, got:\n%s", content)
	}
}

func TestIntegration_CityNotFoundError(t *testing.T) {
	handler := chainWith(
		weather.NewBaseProvider(&failProvider{}),
		weather.NewBaseProvider(&cityNotFoundProvider{}),
		weather.NewBaseProvider(&failProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "Atlantis")
	require.False(t, ok)
	require.ErrorIs(t, err, weather.ErrCityNotFound)
}

func TestIntegration_ChainErrors(t *testing.T) {
	handler := chainWith(
		weather.NewBaseProvider(&failProvider{}),
		weather.NewBaseProvider(&failProvider{}),
	)

	ctx := context.Background()
	_, err := handler.GetWeather(ctx, "Nowhere")
	require.Error(t, err)
	require.Contains(t, err.Error(), "all providers failed")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}
