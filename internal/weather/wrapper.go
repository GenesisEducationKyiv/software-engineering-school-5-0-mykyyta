package weather

import (
	"context"
	"log"
	"time"
)

type Wrapper struct {
	next     Provider
	provider string
	logger   *log.Logger
}

func NewWrapper(next Provider, providerName string, logger *log.Logger) *Wrapper {
	return &Wrapper{
		next:     next,
		provider: providerName,
		logger:   logger,
	}
}

func (p *Wrapper) GetWeather(ctx context.Context, city string) (Report, error) {
	start := time.Now()
	res, err := p.next.GetWeather(ctx, city)
	p.log("GetWeather", city, time.Since(start), err)
	return res, err
}

func (p *Wrapper) CityIsValid(ctx context.Context, city string) (bool, error) {
	start := time.Now()
	ok, err := p.next.CityIsValid(ctx, city)
	p.log("CityIsValid", city, time.Since(start), err)
	return ok, err
}

func (p *Wrapper) log(method, city string, duration time.Duration, err error) {
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
