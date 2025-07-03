package cache

import (
	"context"
	"errors"
	"log"

	"weatherApi/internal/weather"
)

type reader interface {
	Get(ctx context.Context, city, provider string) (weather.Report, error)
}

type metrics interface {
	RecordProviderHit(provider string)
	RecordProviderMiss(provider string)
	RecordTotalHit()
	RecordTotalMiss()
}

type Reader struct {
	Provider      weather.Provider
	Cache         reader
	Metrics       metrics
	ProviderNames []string
}

func NewReader(provider weather.Provider, cache reader, metrics metrics, providerNames []string) Reader {
	return Reader{Provider: provider, Cache: cache, Metrics: metrics, ProviderNames: providerNames}
}

// GetWeather retrieves the weather report by querying multiple cache sources in order.
//
// Cache reads are handled at the Reader level (not inside providers) to enable
// accurate metrics collection. In particular, total cache misses can only be
// detected reliably here, after all sources have been checked.
func (c Reader) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	for _, name := range c.ProviderNames {
		report, err := c.Cache.Get(ctx, city, name)
		if err == nil {
			c.Metrics.RecordProviderHit(name)
			c.Metrics.RecordTotalHit()
			return report, nil
		}
		if errors.Is(err, ErrCacheMiss) {
			c.Metrics.RecordProviderMiss(name)
			continue
		}
		log.Printf("Cache error for %s/%s: %v", city, name, err)
		break
	}

	c.Metrics.RecordTotalMiss()
	return c.Provider.GetWeather(ctx, city)
}

func (c Reader) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
