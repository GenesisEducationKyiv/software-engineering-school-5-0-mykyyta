package chain

import (
	"context"
	"errors"
	"testing"
	"weather/internal/service"

	"weather/internal/domain"

	"github.com/stretchr/testify/require"
)

type MockProvider struct {
	GetWeatherFunc  func(ctx context.Context, city string) (domain.Report, error)
	CityIsValidFunc func(ctx context.Context, city string) (bool, error)
}

func (m *MockProvider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return m.GetWeatherFunc(ctx, city)
}

func (m *MockProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return m.CityIsValidFunc(ctx, city)
}

func ExtractJoinedErrors(err error) []error {
	type multiUnwrapper interface {
		Unwrap() []error
	}

	if err == nil {
		return nil
	}

	if uw, ok := err.(multiUnwrapper); ok {
		return uw.Unwrap()
	}

	return []error{err}
}

func TestBaseProvider_ChainLogic(t *testing.T) {
	ctx := context.Background()

	t.Run("fallback to second", func(t *testing.T) {
		first := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, errors.New("network error")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("network error")
			},
		})

		second := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{Temperature: 25, Description: "Sunny"}, nil
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
		first := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, service.weather.ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, service.weather.ErrCityNotFound
			},
		})
		second := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, service.weather.ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, service.weather.ErrCityNotFound
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Atlantis")
		require.Error(t, err)
		require.True(t, errors.Is(err, service.weather.ErrCityNotFound))

		ok, err := first.CityIsValid(ctx, "Atlantis")
		require.False(t, ok)
		require.Error(t, err)
		require.True(t, errors.Is(err, service.weather.ErrCityNotFound))
	})

	t.Run("mixed errors with not found", func(t *testing.T) {
		first := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, errors.New("timeout")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("timeout")
			},
		})
		second := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, service.weather.ErrCityNotFound
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, service.weather.ErrCityNotFound
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Unknown")
		require.Error(t, err)
		require.True(t, errors.Is(err, service.weather.ErrCityNotFound))

		ok, err := first.CityIsValid(ctx, "Unknown")
		require.False(t, ok)
		require.Error(t, err)
		require.True(t, errors.Is(err, service.weather.ErrCityNotFound))
	})

	t.Run("all fail with non-notfound", func(t *testing.T) {
		first := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, errors.New("bad gateway")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("bad gateway")
			},
		})
		second := NewNode(&MockProvider{
			GetWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, errors.New("rate limit")
			},
			CityIsValidFunc: func(ctx context.Context, city string) (bool, error) {
				return false, errors.New("rate limit")
			},
		})
		first.SetNext(second)

		_, err := first.GetWeather(ctx, "Kyiv")
		require.Error(t, err)
		require.False(t, errors.Is(err, service.weather.ErrCityNotFound))

		ok, err := first.CityIsValid(ctx, "Kyiv")
		require.False(t, ok)
		require.Error(t, err)
		require.False(t, errors.Is(err, service.weather.ErrCityNotFound))

		errs := ExtractJoinedErrors(err)
		require.Len(t, errs, 2)
		require.Contains(t, errs[0].Error(), "bad gateway")
		require.Contains(t, errs[1].Error(), "rate limit")
	})
}
