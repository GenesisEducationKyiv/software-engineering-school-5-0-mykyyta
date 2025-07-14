package cache

import (
	"context"
	"errors"
	"log"
	"time"

	weather "weather/internal/service"

	"weather/internal/domain"
)

type writer interface {
	Set(ctx context.Context, city string, provider string, report domain.Report, ttl time.Duration) error
	SetCityNotFound(ctx context.Context, city, provider string, ttl time.Duration) error
	GetCityNotFound(ctx context.Context, city, provider string) (bool, error)
}

type Writer struct {
	Provider     weather.Provider
	Cache        writer
	ProviderName string
	TTL          time.Duration
	NotFoundTTL  time.Duration
}

func NewWriter(
	provider weather.Provider,
	cache writer,
	providerName string,
	ttl time.Duration,
	notFoundTTL time.Duration,
) Writer {
	return Writer{
		Provider:     provider,
		Cache:        cache,
		ProviderName: providerName,
		TTL:          ttl,
		NotFoundTTL:  notFoundTTL,
	}
}

// GetWeather retrieves the weather report for the given city.
// Before querying the provider, it checks the "city not found" cache.
// If the city is known to be invalid (e.g. a previous request failed),
// it immediately returns ErrCityNotFound.
//
// This is important because some weather providers do not support
// small or less-known cities. Caching negative results avoids repeated
// unnecessary calls to the provider and improves performance.
func (c Writer) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	if notFound, err := c.Cache.GetCityNotFound(ctx, city, c.ProviderName); err == nil && notFound {
		return domain.Report{}, domain.ErrCityNotFound
	} else if err != nil {
		log.Printf("Error checking CityNotFound cache for %q/%s: %v", city, c.ProviderName, err)
	}
	return c.getReportAndCache(ctx, city)
}

func (c Writer) getReportAndCache(ctx context.Context, city string) (domain.Report, error) {
	report, err := c.Provider.GetWeather(ctx, city)
	if err != nil {
		c.cacheCityNotFound(ctx, city, err)
		return report, err
	}
	if cacheErr := c.Cache.Set(ctx, city, c.ProviderName, report, c.TTL); cacheErr != nil {
		log.Printf("Caching weather data for %q/%s: %v", city, c.ProviderName, cacheErr)
	}
	return report, nil
}

func (c Writer) cacheCityNotFound(ctx context.Context, city string, err error) {
	if !errors.Is(err, domain.ErrCityNotFound) {
		return
	}
	if cacheErr := c.Cache.SetCityNotFound(ctx, city, c.ProviderName, c.NotFoundTTL); cacheErr != nil {
		log.Printf("Caching CityNotFound for %q/%s: %v", city, c.ProviderName, cacheErr)
	}
}

func (c Writer) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
