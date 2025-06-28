package cache

import (
	"context"
	"errors"

	"weatherApi/internal/weather"
)

type reader interface {
	Get(ctx context.Context, city, provider string) (weather.Report, error)
}

type metrics interface {
	RecordProvider(provider, status string)
	RecordTotal(status string)
}

type Reader struct {
	Provider      weather.Provider
	Cache         reader
	Metrics       metrics
	ProviderNames []string
}

func NewReader(provider weather.Provider, cache reader, metrics metrics, providerNames []string) *Reader {
	return &Reader{Provider: provider, Cache: cache, Metrics: metrics, ProviderNames: providerNames}
}

func (c *Reader) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	for _, name := range c.ProviderNames {
		report, err := c.Cache.Get(ctx, city, name)
		if err == nil {
			c.Metrics.RecordProvider(name, "hit")
			c.Metrics.RecordTotal("hit")
			return report, nil
		}
		if errors.Is(err, ErrCacheMiss) {
			c.Metrics.RecordProvider(name, "miss")
			continue
		}
		break
	}

	c.Metrics.RecordTotal("miss")
	return c.Provider.GetWeather(ctx, city)
}

func (c *Reader) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
