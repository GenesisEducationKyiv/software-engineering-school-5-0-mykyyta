package weather

import "context"

type Provider interface {
	GetCurrentWeather(ctx context.Context, city string) (*Weather, error)
	CityExists(ctx context.Context, city string) (bool, error)
}

type Service struct {
	provider Provider
}

func NewService(p Provider) *Service {
	return &Service{provider: p}
}

func (s *Service) GetWeather(ctx context.Context, city string) (*Weather, error) {
	return s.provider.GetCurrentWeather(ctx, city)
}

func (s *Service) CheckCity(ctx context.Context, city string) (bool, error) {
	return s.provider.CityExists(ctx, city)
}
