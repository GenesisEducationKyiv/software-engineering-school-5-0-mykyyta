package grpcapi

import (
	"context"

	"weather/internal/domain"
	weatherpb "weather/internal/proto"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Server struct {
	weatherpb.UnimplementedWeatherServiceServer
	service weatherService
}

func NewHandler(s weatherService) *Server {
	return &Server{service: s}
}

func (s *Server) GetWeather(ctx context.Context, req *weatherpb.WeatherRequest) (*weatherpb.WeatherResponse, error) {
	report, err := s.service.GetWeather(ctx, req.City)
	if err != nil {
		return nil, err
	}
	return &weatherpb.WeatherResponse{
		Temperature: report.Temperature,
		Humidity:    int32(report.Humidity),
		Description: report.Description,
	}, nil
}

func (s *Server) ValidateCity(ctx context.Context, req *weatherpb.ValidateRequest) (*weatherpb.ValidateResponse, error) {
	ok, err := s.service.CityIsValid(ctx, req.City)
	if err != nil {
		return nil, err
	}
	return &weatherpb.ValidateResponse{Valid: ok}, nil
}
