package cache

import (
	"context"
	"log"
	"time"

	"weatherApi/internal/weather"
)

type writer interface {
	Set(ctx context.Context, city string, provider string, report weather.Report, ttl time.Duration) error
}

type Writer struct {
	Provider     weather.Provider
	Cache        writer
	ProviderName string
	TTL          time.Duration
}

func NewWriter(provider weather.Provider, cache writer, providerName string, ttl time.Duration) *Writer {
	return &Writer{Provider: provider, Cache: cache, ProviderName: providerName, TTL: ttl}
}

func (c *Writer) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	rep, err := c.Provider.GetWeather(ctx, city)
	if err != nil {
		return rep, err
	}
	if err := c.Cache.Set(ctx, city, c.ProviderName, rep, c.TTL); err != nil {
		log.Printf("Failed to cache weather data for %s/%s: %v", city, c.ProviderName, err)
	}
	return rep, nil
}

func (c *Writer) CityIsValid(ctx context.Context, city string) (bool, error) {
	return c.Provider.CityIsValid(ctx, city)
}
