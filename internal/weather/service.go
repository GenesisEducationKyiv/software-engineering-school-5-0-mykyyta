package weather

import (
	"context"
	"errors"
)

var ErrCityNotFound = errors.New("city not found")

type Provider interface {
	GetWeather(ctx context.Context, city string) (Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Service struct {
	provider Provider
}

func NewService(p Provider) Service {
	return Service{provider: p}
}

func (s Service) GetWeather(ctx context.Context, city string) (Report, error) {
	return s.provider.GetWeather(ctx, city)
}

func (s Service) CityIsValid(ctx context.Context, city string) (bool, error) {
	return s.provider.CityIsValid(ctx, city)
}
