package weatherlogger

import (
	"context"
	"log"
	"time"

	"weatherApi/internal/weather"
)

type Provider struct {
	next     weather.Provider
	provider string
}

func New(next weather.Provider, providerName string) *Provider {
	return &Provider{
		next:     next,
		provider: providerName,
	}
}

func (p *Provider) GetWeather(ctx context.Context, city string) (weather.Report, error) {
	start := time.Now()
	res, err := p.next.GetWeather(ctx, city)
	p.log("GetWeather", city, time.Since(start), err)
	return res, err
}

func (p *Provider) CityIsValid(ctx context.Context, city string) (bool, error) {
	start := time.Now()
	ok, err := p.next.CityIsValid(ctx, city)
	p.log("CityIsValid", city, time.Since(start), err)
	return ok, err
}

func (p *Provider) log(method, city string, duration time.Duration, err error) {
	status := "OK"
	if err != nil {
		status = "ERR: " + err.Error()
	}

	log.Printf("[%s] %s(%q) took %s → %s",
		p.provider,
		method,
		city,
		duration.Truncate(time.Millisecond),
		status,
	)
}
