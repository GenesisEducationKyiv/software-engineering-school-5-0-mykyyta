package service

import (
	"context"
	"fmt"

	"gateway/internal/adapter/subscription"
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
	req.Email = s.securityValidator.SanitizeInput(req.Email)
	req.City = s.securityValidator.SanitizeInput(req.City)
	req.Frequency = s.securityValidator.SanitizeInput(req.Frequency)

	if err := s.securityValidator.ValidateCity(req.City); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Subscribe(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("subscription service failed: %w", err)
	}

	return resp, nil
}

func (s *Service) Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error) {
	if err := s.securityValidator.ValidateToken(token); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Confirm(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("confirm service failed: %w", err)
	}

	return resp, nil
}

func (s *Service) Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error) {
	if err := s.securityValidator.ValidateToken(token); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.Unsubscribe(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("unsubscribe service failed: %w", err)
	}

	return resp, nil
}

func (s *Service) GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error) {
	city = s.securityValidator.SanitizeInput(city)
	if err := s.securityValidator.ValidateCity(city); err != nil {
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	resp, err := s.subscriptionClient.GetWeather(ctx, city)
	if err != nil {
		return nil, fmt.Errorf("weather service failed: %w", err)
	}

	return resp, nil
}
