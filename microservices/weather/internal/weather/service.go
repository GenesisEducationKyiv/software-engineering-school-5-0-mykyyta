package weather

import (
	"context"
	"errors"
	"fmt"

	"weather/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
	logger := loggerPkg.From(ctx)
	logger.Info("getting weather data from provider", "city", city)

	report, err := s.provider.GetWeather(ctx, city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			logger.Warn("city not found in provider", "city", city)
		} else {
			logger.Error("failed to get weather from provider", "city", city, "error", err)
		}
		return domain.Report{}, err
	}

	logger.Info("weather data retrieved from provider",
		"city", city,
		"temperature", report.Temperature,
		"humidity", report.Humidity,
		"description", report.Description)

	return report, nil
}

func (s Service) CityIsValid(ctx context.Context, city string) (bool, error) {
	logger := loggerPkg.From(ctx)
	logger.Info("validating city with provider", "city", city)

	valid, aggErr := s.provider.CityIsValid(ctx, city)
	if aggErr == nil {
		logger.Info("city validation completed by provider", "city", city, "valid", valid)
		return valid, nil
	}

	if errors.Is(aggErr, domain.ErrCityNotFound) {
		logger.Warn("city not found during validation", "city", city)
		return false, domain.ErrCityNotFound
	}

	logger.Error("city validation failed", "city", city, "error", aggErr)
	return false, fmt.Errorf("validation failed for city %q: %w", city, aggErr)
}
