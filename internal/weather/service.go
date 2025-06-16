package weather

import "context"

type weatherProvider interface {
	GetCurrentWeather(ctx context.Context, city string) (Weather, error)
	CityExists(ctx context.Context, city string) (bool, error)
}

type WeatherService struct {
	provider weatherProvider
}

func NewWeatherService(p weatherProvider) *WeatherService {
	return &WeatherService{provider: p}
}

func (s *WeatherService) GetWeather(ctx context.Context, city string) (Weather, error) {
	return s.provider.GetCurrentWeather(ctx, city)
}

func (s *WeatherService) CityIsValid(ctx context.Context, city string) (bool, error) {
	return s.provider.CityExists(ctx, city)
}
