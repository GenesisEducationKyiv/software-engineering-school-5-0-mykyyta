package grpcapi

import (
	"context"
	"errors"

	"weather/internal/domain"
	weatherpb "weather/internal/proto"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Handler struct {
	weatherpb.UnimplementedWeatherServiceServer
	ws weatherService
}

func NewHandler(s weatherService) *Handler {
	return &Handler{ws: s}
}

func (s *Handler) GetWeather(ctx context.Context, req *weatherpb.WeatherRequest) (*weatherpb.WeatherResponse, error) {
	report, err := s.ws.GetWeather(ctx, req.City)
	if err != nil {
		logger := loggerPkg.From(ctx)
		if errors.Is(err, domain.ErrCityNotFound) {
			logger.Warnw("city not found (gRPC)", "city", req.City)
			return nil, err
		}
		logger.Errorw("failed to get weather (gRPC)", "city", req.City, "error", err)
		return nil, err
	}
	return &weatherpb.WeatherResponse{
		Temperature: report.Temperature,
		Humidity:    int32(report.Humidity),
		Description: report.Description,
	}, nil
}

func (s *Handler) ValidateCity(ctx context.Context, req *weatherpb.ValidateRequest) (*weatherpb.ValidateResponse, error) {
	ok, err := s.ws.CityIsValid(ctx, req.City)
	if err != nil {
		logger := loggerPkg.From(ctx)
		if errors.Is(err, domain.ErrCityNotFound) {
			logger.Warnw("city not found (gRPC)", "city", req.City)
			return nil, err
		}
		logger.Errorw("failed to validate city (gRPC)", "city", req.City, "error", err)
		return nil, err
	}

	// Log successful gRPC city validation
	logger := loggerPkg.From(ctx)
	logger.Infow("city validation completed successfully (gRPC)", "city", req.City, "valid", ok)

	return &weatherpb.ValidateResponse{Valid: ok}, nil
}
