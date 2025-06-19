package chain

import (
	"context"
	"errors"
	"fmt"

	"weatherApi/internal/weather"
)

type Provider struct {
	providers []weather.Provider
}

func NewProvider(providers ...weather.Provider) *Provider {
	return &Provider{providers: providers}
}

func (c *Provider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	var lastErr error

	for _, p := range c.providers {
		result, err := p.GetWeather(ctx, city)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return weather.Report{}, fmt.Errorf("all providers failed: %w", lastErr)
}

func (c *Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	var lastErr error
	var cityNotFoundSeen bool

	for _, p := range c.providers {
		ok, err := p.CityIsValid(ctx, city)
		if err == nil {
			return ok, nil
		}
		lastErr = err
		if errors.Is(err, weather.ErrCityNotFound) {
			cityNotFoundSeen = true
		}
	}

	if cityNotFoundSeen {
		return false, weather.ErrCityNotFound
	}
	return false, fmt.Errorf("all providers failed: %w", lastErr)
}
