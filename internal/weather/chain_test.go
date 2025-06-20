package weather

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/net/context"
)

type MockProvider struct {
	Name            string
	GetWeatherFunc  func(ctx context.Context, city string) (Report, error)
	CityIsValidFunc func(ctx context.Context, city string) (bool, error)
}

func (m *MockProvider) GetWeather(ctx context.Context, city string) (Report, error) {
	return m.GetWeatherFunc(ctx, city)
}

func (m *MockProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return m.CityIsValidFunc(ctx, city)
}

func TestChainWeatherProvider_GetWeatherLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("fallback to second", func(t *testing.T) {
		provider := NewChain(
			&MockProvider{
				Name: "broken",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{}, errors.New("network error")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("network error")
				},
			},
			&MockProvider{
				Name: "valid",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{Temperature: 25, Description: "Sunny"}, nil
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return true, nil
				},
			},
		)

		w, err := provider.GetWeather(ctx, "Kyiv")
		require.NoError(t, err)
		require.Equal(t, "Sunny", w.Description)

		ok, err := provider.CityIsValid(ctx, "Kyiv")
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("all not found", func(t *testing.T) {
		notFound := &MockProvider{
			Name: "notFound",
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, ErrCityNotFound
			},
		}
		provider := NewChain(notFound, notFound)

		_, err := provider.GetWeather(ctx, "Atlantis")
		require.ErrorIs(t, err, ErrCityNotFound)

		ok, err := provider.CityIsValid(ctx, "Atlantis")
		require.False(t, ok)
		require.ErrorIs(t, err, ErrCityNotFound)
	})

	t.Run("mixed errors with not found", func(t *testing.T) {
		provider := NewChain(
			&MockProvider{
				Name: "timeout",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{}, errors.New("timeout")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("timeout")
				},
			},
			&MockProvider{
				Name: "notFound",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{}, ErrCityNotFound
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, ErrCityNotFound
				},
			},
		)

		_, err := provider.GetWeather(ctx, "Unknown")
		require.ErrorIs(t, err, ErrCityNotFound)

		ok, err := provider.CityIsValid(ctx, "Unknown")
		require.False(t, ok)
		require.ErrorIs(t, err, ErrCityNotFound)
	})

	t.Run("all fail with non-notfound", func(t *testing.T) {
		provider := NewChain(
			&MockProvider{
				Name: "bad gateway",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{}, errors.New("bad gateway")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("bad gateway")
				},
			},
			&MockProvider{
				Name: "rate limit",
				GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
					return Report{}, errors.New("rate limit")
				},
				CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
					return false, errors.New("rate limit")
				},
			},
		)

		_, err := provider.GetWeather(ctx, "Kyiv")
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrCityNotFound)

		ok, err := provider.CityIsValid(ctx, "Kyiv")
		require.False(t, ok)
		require.NotErrorIs(t, err, ErrCityNotFound)
	})
}
