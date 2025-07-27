package service

import (
	"context"
	"fmt"

	"gateway/internal/adapter/subscription"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type SecurityValidator interface {
	ValidateToken(token string) error
	ValidateCity(city string) error
	SanitizeInput(input string) string
}

type SubscriptionClient interface {
	Subscribe(ctx context.Context, req subscription.SubscribeRequest) (*subscription.SubscribeResponse, error)
	Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error)
	Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error)
	GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error)
}

type Service struct {
	subscriptionClient SubscriptionClient
	securityValidator  SecurityValidator
}

func NewService(
	subscriptionClient SubscriptionClient,
	securityValidator SecurityValidator,
) *Service {
	return &Service{
		subscriptionClient: subscriptionClient,
		securityValidator:  securityValidator,
	}
}

func (s *Service) Subscribe(ctx context.Context, req subscription.SubscribeRequest) (*subscription.SubscribeResponse, error) {
	logger := loggerPkg.From(ctx)

	req.Email = s.securityValidator.SanitizeInput(req.Email)
	req.City = s.securityValidator.SanitizeInput(req.City)
	req.Frequency = s.securityValidator.SanitizeInput(req.Frequency)

	if err := s.securityValidator.ValidateCity(req.City); err != nil {
		logger.Warnw("Security validation failed", "validation_error", err, "city", req.City)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Subscribe(ctx, req)
	if err != nil {
		logger.Errorw("Subscription service call failed", "err", err, "city", req.City)
		return nil, fmt.Errorf("subscription service failed: %w", err)
	}

	logger.Debugw("Subscription successful", "city", req.City)
	return resp, nil
}

func (s *Service) Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error) {
	logger := loggerPkg.From(ctx)

	if err := s.securityValidator.ValidateToken(token); err != nil {
		logger.Warnw("Token validation failed", "validation_error", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Confirm(ctx, token)
	if err != nil {
		logger.Errorw("Confirm service call failed", "err", err)
		return nil, fmt.Errorf("confirm service failed: %w", err)
	}

	logger.Debugw("Confirmation successful")
	return resp, nil
}

func (s *Service) Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error) {
	logger := loggerPkg.From(ctx)

	if err := s.securityValidator.ValidateToken(token); err != nil {
		logger.Warnw("Token validation failed", "validation_error", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Unsubscribe(ctx, token)
	if err != nil {
		logger.Errorw("Unsubscribe service call failed", "err", err)
		return nil, fmt.Errorf("unsubscribe service failed: %w", err)
	}

	logger.Debugw("Unsubscribe successful")
	return resp, nil
}

func (s *Service) GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error) {
	logger := loggerPkg.From(ctx)

	city = s.securityValidator.SanitizeInput(city)
	if err := s.securityValidator.ValidateCity(city); err != nil {
		logger.Warnw("City validation failed", "validation_error", err, "city", city)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.GetWeather(ctx, city)
	if err != nil {
		logger.Errorw("Weather service call failed", "err", err, "city", city)
		return nil, fmt.Errorf("weather service failed: %w", err)
	}

	logger.Debugw("Weather fetch successful", "city", city)
	return resp, nil
}
