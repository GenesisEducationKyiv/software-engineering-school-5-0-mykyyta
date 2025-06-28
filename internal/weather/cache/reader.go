package cache

import (
	"context"
	"errors"

	"weatherApi/internal/weather/metrics"

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
	for _, providerName := range c.ProviderNames {
		report, err := c.Cache.Get(ctx, city, providerName)
		if err == nil {
			metrics.CacheHits.WithLabelValues(providerName).Inc()
			metrics.CacheResult.WithLabelValues("hit").Inc()
			return report, nil
		}
		if errors.Is(err, ErrCacheMiss) {
			metrics.CacheMisses.WithLabelValues(providerName).Inc()
			continue
		}
		break
	}
	metrics.CacheResult.WithLabelValues("miss").Inc()
	return c.Provider.GetWeather(ctx, city)
}

func (c *Reader) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
