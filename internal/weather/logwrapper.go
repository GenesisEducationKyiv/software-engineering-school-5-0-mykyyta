package weather

import (
	"context"
	"log"
	"time"
)

type LogWrapper struct {
	next     Provider
	provider string
	logger   *log.Logger
}

func NewLogWrapper(next Provider, providerName string, logger *log.Logger) LogWrapper {
	return LogWrapper{
		next:     next,
		provider: providerName,
		logger:   logger,
	}
}

func (p LogWrapper) GetWeather(ctx context.Context, city string) (Report, error) {
	start := time.Now()
	res, err := p.next.GetWeather(ctx, city)
	p.log("GetWeather", city, time.Since(start), err)
	return res, err
}

func (p LogWrapper) CityIsValid(ctx context.Context, city string) (bool, error) {
	start := time.Now()
	ok, err := p.next.CityIsValid(ctx, city)
	p.log("CityIsValid", city, time.Since(start), err)
	return ok, err
}

func (p LogWrapper) log(method, city string, duration time.Duration, err error) {
	status := "OK"
	if err != nil {
		status = "ERR: " + err.Error()
	}

	p.logger.Printf("[%s] %s(%q) took %s → %s",
		p.provider,
		method,
		city,
		duration.Truncate(time.Millisecond),
		status,
	)
}
