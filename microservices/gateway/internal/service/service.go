package service

import (
	"context"
	"fmt"
	"gateway/internal/adapter/subscription"
	"log"
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
	logger             *log.Logger
}

func NewService(
	subscriptionClient SubscriptionClient,
	securityValidator SecurityValidator,
	logger *log.Logger,
) *Service {
	return &Service{
		subscriptionClient: subscriptionClient,
		securityValidator:  securityValidator,
		logger:             logger,
	}
}

func (s *Service) Subscribe(ctx context.Context, req subscription.SubscribeRequest) (*subscription.SubscribeResponse, error) {
	req.Email = s.securityValidator.SanitizeInput(req.Email)
	req.City = s.securityValidator.SanitizeInput(req.City)
	req.Frequency = s.securityValidator.SanitizeInput(req.Frequency)

	if err := s.securityValidator.ValidateCity(req.City); err != nil {
		s.logger.Printf("Security validation failed for city: %v", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	s.logger.Printf("Processing subscribe request for email: %s, city: %s", req.Email, req.City)

	resp, err := s.subscriptionClient.Subscribe(ctx, req)
	if err != nil {
		s.logger.Printf("Subscription service failed: %v", err)
		return nil, fmt.Errorf("subscription service failed: %w", err)
	}

	s.logger.Printf("Successfully processed subscribe request for: %s", req.Email)
	return resp, nil
}

func (s *Service) Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error) {
	if err := s.securityValidator.ValidateToken(token); err != nil {
		s.logger.Printf("Token security validation failed: %v", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	s.logger.Printf("Processing confirm request for token: %s", token)

	resp, err := s.subscriptionClient.Confirm(ctx, token)
	if err != nil {
		s.logger.Printf("Confirm service failed: %v", err)
		return nil, fmt.Errorf("confirm service failed: %w", err)
	}

	s.logger.Printf("Successfully processed confirm request for token: %s", token)
	return resp, nil
}

func (s *Service) Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error) {
	if err := s.securityValidator.ValidateToken(token); err != nil {
		s.logger.Printf("Token security validation failed: %v", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	s.logger.Printf("Processing unsubscribe request for token: %s", token)

	resp, err := s.subscriptionClient.Unsubscribe(ctx, token)
	if err != nil {
		s.logger.Printf("Unsubscribe service failed: %v", err)
		return nil, fmt.Errorf("unsubscribe service failed: %w", err)
	}

	s.logger.Printf("Successfully processed unsubscribe request for token: %s", token)
	return resp, nil
}

func (s *Service) GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error) {
	city = s.securityValidator.SanitizeInput(city)
	if err := s.securityValidator.ValidateCity(city); err != nil {
		s.logger.Printf("City security validation failed: %v", err)
		return nil, fmt.Errorf("security validation failed: %w", err)
	}

	s.logger.Printf("Processing weather request for city: %s", city)

	resp, err := s.subscriptionClient.GetWeather(ctx, city)
	if err != nil {
		s.logger.Printf("Weather service failed for city %s: %v", city, err)
		return nil, fmt.Errorf("weather service failed: %w", err)
	}

	s.logger.Printf("Successfully processed weather request for city: %s", city)
	return resp, nil
}
