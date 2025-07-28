package logger

import (
	"context"
	"time"

	"weather/internal/weather"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"weather/internal/domain"
)

type LogWrapper struct {
	next     weather.Provider
	provider string
}

func NewWrapper(next weather.Provider, providerName string) LogWrapper {
	return LogWrapper{
		next:     next,
		provider: providerName,
	}
}

func (p LogWrapper) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	start := time.Now()
	res, err := p.next.GetWeather(ctx, city)
	dur := time.Since(start)
	status := "OK"
	if err != nil {
		status = err.Error()
	}
	logger := loggerPkg.From(ctx)
	logger.Info(
		"provider call",
		"provider", p.provider,
		"method", "GetWeather",
		"city", city,
		"duration_ms", dur.Milliseconds(),
		"status", status,
	)
	return res, err
}

func (p LogWrapper) CityIsValid(ctx context.Context, city string) (bool, error) {
	start := time.Now()
	ok, err := p.next.CityIsValid(ctx, city)
	dur := time.Since(start)
	status := "OK"
	if err != nil {
		status = err.Error()
	}
	logger := loggerPkg.From(ctx)
	logger.Info(
		"provider call",
		"provider", p.provider,
		"method", "CityIsValid",
		"city", city,
		"duration_ms", dur.Milliseconds(),
		"status", status,
	)
	return ok, err
}
