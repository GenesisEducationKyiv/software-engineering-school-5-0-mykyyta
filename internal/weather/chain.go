package weather

import (
	"errors"
	"fmt"
	"golang.org/x/net/context"
)

type ChainWeatherProvider struct {
	providers []WeatherProvider
}

func NewChainWeatherProvider(providers ...WeatherProvider) *ChainWeatherProvider {
	return &ChainWeatherProvider{providers: providers}
}

func (c *ChainWeatherProvider) GetWeather(ctx context.Context, city string) (Weather, error) {
	return tryProviders(c.providers, func(p WeatherProvider) (Weather, error) {
		return p.GetWeather(ctx, city)
	})
}

func (c *ChainWeatherProvider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return tryProviders(c.providers, func(p WeatherProvider) (bool, error) {
		return p.CityIsValid(ctx, city)
	})
}

func tryProviders[T any](
	providers []WeatherProvider,
	call func(WeatherProvider) (T, error),
) (T, error) {
	var lastErr error
	var cityNotFoundSeen bool

	for _, p := range providers {
		result, err := call(p)
		if err == nil {
			return result, nil
		}
		lastErr = err
		if errors.Is(err, ErrCityNotFound) {
			cityNotFoundSeen = true
		}
	}

	var zero T

	if cityNotFoundSeen {
		return zero, ErrCityNotFound
	}
	return zero, fmt.Errorf("all providers failed: %w", lastErr)
}
