package weather

import (
	"context"
)

type WeatherProvider interface {
	GetWeather(ctx context.Context, city string) (Weather, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type WeatherService struct {
	provider WeatherProvider
}

func NewWeatherService(p WeatherProvider) *WeatherService {
	return &WeatherService{provider: p}
}

func (s *WeatherService) GetWeather(ctx context.Context, city string) (Weather, error) {
	return s.provider.GetWeather(ctx, city)
}

func (s *WeatherService) CityIsValid(ctx context.Context, city string) (bool, error) {
	return s.provider.CityIsValid(ctx, city)
}
