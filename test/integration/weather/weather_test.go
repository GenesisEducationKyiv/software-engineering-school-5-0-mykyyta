//go:build integration

package weather_test

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"weatherApi/internal/domain"
	"weatherApi/internal/weather"
	"weatherApi/internal/weather/cache"
	"weatherApi/internal/weather/chain"
	"weatherApi/internal/weather/logger"

	redismock "github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/require"
)

// --- Fake Providers ---

type successProvider struct{}

func (s *successProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return domain.Report{Temperature: 20.5, Humidity: 60, Description: "Sunny"}, nil
}

func (s *successProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}

type failProvider struct{}

func (f *failProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return domain.Report{}, context.DeadlineExceeded
}

func (f *failProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, context.DeadlineExceeded
}

type cityNotFoundProvider struct{}

func (c *cityNotFoundProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return domain.Report{}, weather.ErrCityNotFound
}

func (c *cityNotFoundProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return false, weather.ErrCityNotFound
}

// --- Test Utils ---

func chainWith(handlers ...chain.ChainableProvider) chain.ChainableProvider {
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
	logg := log.New(file, "", log.LstdFlags)

	fp := &failProvider{}
	sp := &successProvider{}
	wrappedFail := logger.NewWrapper(fp, "FailProvider", logg)
	wrappedSuccess := logger.NewWrapper(sp, "SuccessProvider", logg)

	handler := chainWith(
		chain.NewNode(wrappedFail),
		chain.NewNode(wrappedSuccess),
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
		chain.NewNode(&failProvider{}),
		chain.NewNode(&cityNotFoundProvider{}),
		chain.NewNode(&failProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "Atlantis")
	require.False(t, ok)
	require.ErrorIs(t, err, weather.ErrCityNotFound)
}

func TestIntegration_ChainErrors(t *testing.T) {
	handler := chainWith(
		chain.NewNode(&failProvider{}),
		chain.NewNode(&failProvider{}),
	)

	ctx := context.Background()
	_, err := handler.GetWeather(ctx, "Nowhere")
	require.Error(t, err)
	require.Contains(t, err.Error(), "all providers failed")
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestIntegration_CityIsValid_PrioritizeNotFound(t *testing.T) {
	handler := chainWith(
		chain.NewNode(&failProvider{}),
		chain.NewNode(&cityNotFoundProvider{}),
		chain.NewNode(&failProvider{}),
	)

	ctx := context.Background()
	ok, err := handler.CityIsValid(ctx, "UnknownCity")

	require.False(t, ok)
	require.ErrorIs(t, err, weather.ErrCityNotFound)
}

func TestIntegration_CityIsValid_SkipsCityNotFoundIfLaterSucceeds(t *testing.T) {
	handler := chainWith(
		chain.NewNode(&failProvider{}),
		chain.NewNode(&cityNotFoundProvider{}),
		chain.NewNode(&successProvider{}),
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
	provider := "WeatherAPI"
	normalizedCity := "kyiv"
	cacheKey := "weather:report:" + normalizedCity + ":" + provider

	expected := domain.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}

	data, err := json.Marshal(expected)
	require.NoError(t, err)

	mock.ExpectGet(cacheKey).SetVal(string(data))

	redisCache := cache.NewRedisCache(db)
	reader := cache.NewReader(&failProvider{}, redisCache, &nopMetrics{}, []string{provider})

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, expected, report)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIntegration_CacheReader_Miss(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := context.Background()
	city := "Kyiv"
	provider := "WeatherAPI"
	normalizedCity := "kyiv"
	cacheKey := "weather:report:" + normalizedCity + ":" + provider

	mock.ExpectGet(cacheKey).RedisNil()

	redisCache := cache.NewRedisCache(db)
	reader := cache.NewReader(&successProvider{}, redisCache, &nopMetrics{}, []string{provider})

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, "Sunny", report.Description)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestIntegration_CacheWriter_WritesToRedis_And_ReaderReadsIt(t *testing.T) {
	db, mock := redismock.NewClientMock()
	defer db.Close()

	ctx := context.Background()
	city := "Kyiv"
	provider := "WeatherAPI"
	normalizedCity := "kyiv"
	reportKey := "weather:report:" + normalizedCity + ":" + provider
	notFoundKey := "weather:notfound:" + normalizedCity + ":" + provider
	ttl := 5 * time.Minute
	notFoundTTL := 12 * time.Hour

	expected := domain.Report{
		Temperature: 20.5,
		Humidity:    60,
		Description: "Sunny",
	}

	data, err := json.Marshal(expected)
	require.NoError(t, err)

	mock.ExpectGet(notFoundKey).RedisNil()
	mock.ExpectSet(reportKey, data, ttl).SetVal("OK")
	mock.ExpectGet(reportKey).SetVal(string(data))

	redisCache := cache.NewRedisCache(db)
	writer := cache.NewWriter(&successProvider{}, redisCache, provider, ttl, notFoundTTL)

	_, err = writer.GetWeather(ctx, city)
	require.NoError(t, err)

	reader := cache.NewReader(&failProvider{}, redisCache, &nopMetrics{}, []string{provider})

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, expected, report)

	require.NoError(t, mock.ExpectationsWereMet())
}

type readerStub struct {
	CalledKeys []string
	Responses  map[string]struct {
		Report domain.Report
		Err    error
	}
}

func (r *readerStub) Get(ctx context.Context, city, provider string) (domain.Report, error) {
	key := provider
	r.CalledKeys = append(r.CalledKeys, key)

	resp, ok := r.Responses[key]
	if !ok {
		return domain.Report{}, cache.ErrCacheMiss
	}
	return resp.Report, resp.Err
}

func TestIntegration_Reader_TriesProvidersInOrder(t *testing.T) {
	ctx := context.Background()
	providerNames := []string{"CacheA", "CacheB", "CacheC"}

	expected := domain.Report{Temperature: 22, Humidity: 55, Description: "Cloudy"}

	stub := &readerStub{
		Responses: map[string]struct {
			Report domain.Report
			Err    error
		}{
			"CacheA": {Err: cache.ErrCacheMiss},
			"CacheB": {Report: expected, Err: nil},
		},
	}

	reader := cache.NewReader(&failProvider{}, stub, &nopMetrics{}, providerNames)

	report, err := reader.GetWeather(ctx, "Kyiv")
	require.NoError(t, err)
	require.Equal(t, expected, report)

	require.Equal(t, []string{"CacheA", "CacheB"}, stub.CalledKeys)
}

func TestIntegration_Reader_StopsOnRedisError(t *testing.T) {
	ctx := context.Background()
	city := "Kyiv"
	providerNames := []string{"CacheA", "CacheB", "CacheC"}

	stub := &readerStub{
		Responses: map[string]struct {
			Report domain.Report
			Err    error
		}{
			"CacheA": {Err: cache.ErrCacheMiss},
			"CacheB": {Err: errors.New("connection lost")},
			"CacheC": {Report: domain.Report{Temperature: 99}, Err: nil},
		},
	}

	reader := cache.NewReader(&successProvider{}, stub, &nopMetrics{}, providerNames)

	report, err := reader.GetWeather(ctx, city)
	require.NoError(t, err)
	require.Equal(t, "Sunny", report.Description)

	require.Equal(t, []string{"CacheA", "CacheB"}, stub.CalledKeys)
}
