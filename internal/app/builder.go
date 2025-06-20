package app

import (
	"log"

	"weatherApi/internal/auth"
	"weatherApi/internal/config"
	"weatherApi/internal/db"
	"weatherApi/internal/email"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"
	"weatherApi/internal/weather/providers/tomorrowio"
	"weatherApi/internal/weather/providers/weatherapi"
)

type ProviderSet struct {
	EmailProvider   email.EmailProvider
	TokenProvider   auth.TokenProvider
	WeatherProvider weather.Provider
}
type ServiceSet struct {
	SubService     *subscription.SubscriptionService
	WeatherService *weather.Service
	EmailService   *email.EmailService
}

func BuildProviders(cfg *config.Config, logger *log.Logger) *ProviderSet {
	emailProvider := email.NewSendGridProvider(cfg.SendGridKey, cfg.EmailFrom)
	tokenProvider := auth.NewJWTService(cfg.JWTSecret)

	weatherP1 := weather.NewWrapper(weatherapi.New(cfg.WeatherAPIKey), "WeatherAPI", logger)
	weatherP2 := weather.NewWrapper(tomorrowio.New(cfg.TomorrowioAPIKey), "TomorrowIO", logger)
	weatherProvider := weather.NewChain(weatherP1, weatherP2)

	return &ProviderSet{
		EmailProvider:   emailProvider,
		TokenProvider:   tokenProvider,
		WeatherProvider: weatherProvider,
	}
}

func BuildServices(db *db.DB, cfg *config.Config, p *ProviderSet) *ServiceSet {
	emailService := email.NewEmailService(p.EmailProvider, cfg.BaseURL)
	tokenService := auth.NewTokenService(p.TokenProvider)
	weatherService := weather.NewService(p.WeatherProvider)

	subRepo := subscription.NewSubscriptionRepository(db.Gorm)
	subService := subscription.NewSubscriptionService(subRepo, emailService, weatherService, tokenService)

	return &ServiceSet{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}
}
