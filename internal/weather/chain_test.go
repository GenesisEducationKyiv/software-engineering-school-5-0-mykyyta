package weather_test

import (
	"errors"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
	"testing"
	"weatherApi/internal/weather"
)

type MockProvider struct {
	Name            string
	GetWeatherFunc  func(ctx context.Context, city string) (weather.Weather, error)
	CityIsValidFunc func(ctx context.Context, city string) (bool, error)
}

func (m *MockProvider) GetWeather(ctx context.Context, city string) (weather.Weather, error) {
	return m.GetWeatherFunc(ctx, city)
}

func (m *MockProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return m.CityIsValidFunc(ctx, city)
}

func TestChainWeatherProvider_GetWeatherLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("fallback to second", func(t *testing.T) {
		chain := weather.NewChainWeatherProvider(
			&MockProvider{
				Name: "broken",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{}, errors.New("network error")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("network error")
				},
			},
			&MockProvider{
				Name: "valid",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{Temperature: 25, Description: "Sunny"}, nil
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return true, nil
				},
			},
		)

		w, err := chain.GetWeather(ctx, "Kyiv")
		require.NoError(t, err)
		require.Equal(t, "Sunny", w.Description)

		ok, err := chain.CityIsValid(ctx, "Kyiv")
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("all not found", func(t *testing.T) {
		notFound := &MockProvider{
			Name: "notFound",
			GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
				return weather.Weather{}, weather.ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, weather.ErrCityNotFound
			},
		}
		chain := weather.NewChainWeatherProvider(notFound, notFound)

		_, err := chain.GetWeather(ctx, "Atlantis")
		require.ErrorIs(t, err, weather.ErrCityNotFound)

		ok, err := chain.CityIsValid(ctx, "Atlantis")
		require.False(t, ok)
		require.ErrorIs(t, err, weather.ErrCityNotFound)
	})

	t.Run("mixed errors with not found", func(t *testing.T) {
		chain := weather.NewChainWeatherProvider(
			&MockProvider{
				Name: "timeout",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{}, errors.New("timeout")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("timeout")
				},
			},
			&MockProvider{
				Name: "notFound",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{}, weather.ErrCityNotFound
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, weather.ErrCityNotFound
				},
			},
		)

		_, err := chain.GetWeather(ctx, "Unknown")
		require.ErrorIs(t, err, weather.ErrCityNotFound)

		ok, err := chain.CityIsValid(ctx, "Unknown")
		require.False(t, ok)
		require.ErrorIs(t, err, weather.ErrCityNotFound)
	})

	t.Run("all fail with non-notfound", func(t *testing.T) {
		chain := weather.NewChainWeatherProvider(
			&MockProvider{
				Name: "bad gateway",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{}, errors.New("bad gateway")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("bad gateway")
				},
			},
			&MockProvider{
				Name: "rate limit",
				GetWeatherFunc: func(ctx context.Context, city string) (weather.Weather, error) {
					return weather.Weather{}, errors.New("rate limit")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("rate limit")
				},
			},
		)

		_, err := chain.GetWeather(ctx, "Kyiv")
		require.Error(t, err)
		require.NotErrorIs(t, err, weather.ErrCityNotFound)

		ok, err := chain.CityIsValid(ctx, "Kyiv")
		require.False(t, ok)
		require.NotErrorIs(t, err, weather.ErrCityNotFound)
	})
}
