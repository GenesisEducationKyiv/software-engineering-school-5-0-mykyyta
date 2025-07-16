package weather

import (
	"context"
	"errors"
	"fmt"

	"weather/internal/domain"
)

type Provider interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Service struct {
	provider Provider
}

func NewService(p Provider) Service {
	return Service{provider: p}
}

func (s Service) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return s.provider.GetWeather(ctx, city)
}

func (s Service) CityIsValid(ctx context.Context, city string) (bool, error) {
	valid, aggErr := s.provider.CityIsValid(ctx, city)
	if aggErr == nil {
		return valid, nil
	}

	if errors.Is(aggErr, domain.ErrCityNotFound) {
		return false, domain.ErrCityNotFound
	}

	return false, fmt.Errorf("validation failed for city %q: %w", city, aggErr)
}
