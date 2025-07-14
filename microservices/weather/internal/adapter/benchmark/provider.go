package benchmark

import (
	"context"
	"math/rand"
	"time"
	"weather/internal/domain"
)

type Provider struct{}

func NewProvider() *Provider {
	return &Provider{}
}

func (m *Provider) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	delay := time.Duration(rand.Intn(800)+200) * time.Millisecond
	time.Sleep(delay)

	return domain.Report{
		Temperature: 21.0,
		Humidity:    60,
		Description: "benchmark-mock",
	}, nil
}

func (m *Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	return true, nil
}
