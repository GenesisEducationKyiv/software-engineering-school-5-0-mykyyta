package weather

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type MockProvider struct {
	GetWeatherFunc  func(ctx context.Context, city string) (Report, error)
	CityIsValidFunc func(ctx context.Context, city string) (bool, error)
}

func (m *MockProvider) GetWeather(ctx context.Context, city string) (Report, error) {
	return m.GetWeatherFunc(ctx, city)
}

func (m *MockProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return m.CityIsValidFunc(ctx, city)
}

func TestBaseProvider_ChainLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("fallback to second", func(t *testing.T) {
		first := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, errors.New("network error")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("network error")
			},
		})

		second := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{Temperature: 25, Description: "Sunny"}, nil
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return true, nil
			},
		})

		first.SetNext(second)

		res, err := first.GetWeather(ctx, "Kyiv")
		require.NoError(t, err)
		require.Equal(t, "Sunny", res.Description)

		ok, err := first.CityIsValid(ctx, "Kyiv")
		require.NoError(t, err)
		require.True(t, ok)
	})

	t.Run("all not found", func(t *testing.T) {
		first := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, ErrCityNotFound
			},
		})
		second := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, ErrCityNotFound
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Atlantis")
		require.ErrorIs(t, err, ErrCityNotFound)

		ok, err := first.CityIsValid(ctx, "Atlantis")
		require.False(t, ok)
		require.ErrorIs(t, err, ErrCityNotFound)
	})

	t.Run("mixed errors with not found", func(t *testing.T) {
		first := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, errors.New("timeout")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("timeout")
			},
		})
		second := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, ErrCityNotFound
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Unknown")
		require.ErrorIs(t, err, ErrCityNotFound)

		ok, err := first.CityIsValid(ctx, "Unknown")
		require.False(t, ok)
		require.ErrorIs(t, err, ErrCityNotFound)
	})

	t.Run("all fail with non-notfound", func(t *testing.T) {
		first := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, errors.New("bad gateway")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("bad gateway")
			},
		})
		second := NewBase(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (Report, error) {
				return Report{}, errors.New("rate limit")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("rate limit")
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Kyiv")
		require.Error(t, err)
		require.NotErrorIs(t, err, ErrCityNotFound)

		ok, err := first.CityIsValid(ctx, "Kyiv")
		require.False(t, ok)
		require.NotErrorIs(t, err, ErrCityNotFound)
	})
}
