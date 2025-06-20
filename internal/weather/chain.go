package weather

import (
	"context"
	"errors"
	"fmt"
)

type Chain struct {
	providers []Provider
}

func NewChain(providers ...Provider) *Chain {
	return &Chain{providers: providers}
}

func (c *Chain) GetWeather(ctx context.Context, city string) (Report, error) {
	var lastErr error

	for _, p := range c.providers {
		result, err := p.GetWeather(ctx, city)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	return Report{}, fmt.Errorf("all providers failed: %w", lastErr)
}

func (c *Chain) CityIsValid(ctx context.Context, city string) (bool, error) {
	var lastErr error
	var cityNotFoundSeen bool

	for _, p := range c.providers {
		ok, err := p.CityIsValid(ctx, city)
		if err == nil {
			return ok, nil
		}
		lastErr = err
		if errors.Is(err, ErrCityNotFound) {
			cityNotFoundSeen = true
		}
	}

	if cityNotFoundSeen {
		return false, ErrCityNotFound
	}
	return false, fmt.Errorf("all providers failed: %w", lastErr)
}
