package cache

import (
	"context"
	"errors"

	"weatherApi/internal/weather"
)

type reader interface {
	Get(ctx context.Context, city, provider string) (weather.Report, error)
}

type Reader struct {
	Provider      weather.Provider
	Cache         reader
	ProviderNames []string
}

func NewReader(provider weather.Provider, cache reader, providerNames []string) *Reader {
	return &Reader{Provider: provider, Cache: cache, ProviderNames: providerNames}
}

func (c *Reader) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	for _, name := range c.ProviderNames {
		report, err := c.Cache.Get(ctx, city, name)
		if err == nil {
			return report, nil
		}
		if !errors.Is(err, ErrCacheMiss) {
			break
		}
	}
	return c.Provider.GetWeather(ctx, city)
}

func (c *Reader) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
