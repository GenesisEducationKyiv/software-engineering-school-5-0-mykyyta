//go:build integration

package weather_test

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"weatherApi/internal/weather"
	"weatherApi/internal/weather/cache"

	redismock "github.com/go-redis/redismock/v9"
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

func chainWith(handlers ...weather.ChainableProvider) weather.ChainableProvider {
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
	wrappedFail := weather.NewLogWrapper(fp, "FailProvider", logger)
	wrappedSuccess := weather.NewLogWrapper(sp, "SuccessProvider", logger)

	handler := chainWith(
		weather.NewChainNode(wrappedFail),
		weather.NewChainNode(wrappedSuccess),
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
		weather.NewChainNode(&failProvider{}),
		weather.NewChainNode(&cityNotFoundProvider{}),
		weather.NewChainNode(&failProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "Atlantis")
	require.False(t, ok)
	require.ErrorIs(t, err, weather.ErrCityNotFound)
}

func TestIntegration_ChainErrors(t *testing.T) {
	handler := chainWith(
		weather.NewChainNode(&failProvider{}),
		weather.NewChainNode(&failProvider{}),
	)

	ctx := context.Background()
	_, err := handler.GetWeather(ctx, "Nowhere")
	require.Error(t, err)
	require.Contains(t, err.Error(), "all providers failed")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestIntegration_CityIsValid_PrioritizeNotFound(t *testing.T) {
	handler := chainWith(
		weather.NewChainNode(&failProvider{}),
		weather.NewChainNode(&cityNotFoundProvider{}),
		weather.NewChainNode(&failProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "UnknownCity")

	require.False(t, ok)
	require.ErrorIs(t, err, weather.ErrCityNotFound)
}

func TestIntegration_CityIsValid_SkipsCityNotFoundIfLaterSucceeds(t *testing.T) {
	handler := chainWith(
		weather.NewChainNode(&failProvider{}),
		weather.NewChainNode(&cityNotFoundProvider{}),
		weather.NewChainNode(&successProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "Kyiv")

	require.True(t, ok)
	require.NoError(t, err)
}

// --- CACHE TESTS ---
type nopMetrics struct{}

func (n *nopMetrics) RecordProviderHit(provider string)  {}
func (n *nopMetrics) RecordProviderMiss(provider string) {}
func (n *nopMetrics) RecordTotalHit()                    {}
func (n *nopMetrics) RecordTotalMiss()                   {}

func TestIntegration_CacheReader_Hit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := context.Background()
	city := "Kyiv"
	cacheKey := "weather:Kyiv:WeatherAPI"

	expected := weather.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}

	payload := `{"temperature":20.5,"humidity":60,"description":"Sunny"}`

	mock.ExpectGet(cacheKey).SetVal(payload)

	provider := &failProvider{}
	redisCache := cache.NewRedisCache(db)

	metrics := &nopMetrics{}

	reader := cache.NewReader(provider, redisCache, metrics, []string{"WeatherAPI"})

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, expected.Description, report.Description)
	require.Equal(t, expected.Humidity, report.Humidity)
	require.Equal(t, expected.Temperature, report.Temperature)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

func TestIntegration_CacheReader_Miss(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := context.Background()
	city := "Kyiv"
	cacheKey := "weather:Kyiv:WeatherAPI"

	mock.ExpectGet(cacheKey).RedisNil() // імітуємо промах

	provider := &successProvider{}
	redisCache := cache.NewRedisCache(db)
	metrics := &nopMetrics{}

	reader := cache.NewReader(provider, redisCache, metrics, []string{"WeatherAPI"})

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, "Sunny", report.Description)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}

type mockWriter struct {
	calledWith struct {
		city     string
		provider string
		report   weather.Report
		ttl      time.Duration
	}
	called bool
}

func (m *mockWriter) Set(ctx context.Context, city string, provider string, report weather.Report, ttl time.Duration) error {
	m.calledWith.city = city
	m.calledWith.provider = provider
	m.calledWith.report = report
	m.calledWith.ttl = ttl
	m.called = true
	return nil
}

func TestIntegration_CacheWriter_WritesToRedis(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := context.Background()
	city := "Kyiv"
	cacheKey := "weather:Kyiv:WeatherAPI"

	expected := weather.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}

	data, err := json.Marshal(expected)
	require.NoError(t, err)

	ttl := 300 * time.Second

	mock.ExpectSet(cacheKey, data, ttl).SetVal("OK")

	provider := &successProvider{}
	redisCache := cache.NewRedisCache(db)

	writer := cache.NewWriter(provider, redisCache, "WeatherAPI", ttl)

	report, err := writer.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, expected.Description, report.Description)

	err = mock.ExpectationsWereMet()
	require.NoError(t, err)
}
