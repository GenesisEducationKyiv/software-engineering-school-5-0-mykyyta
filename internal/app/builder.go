package app

import (
	"log"

	"weatherApi/internal/config"
	"weatherApi/internal/db"
	"weatherApi/internal/email"
	"weatherApi/internal/subscription"
	"weatherApi/internal/token"
	"weatherApi/internal/weather"
	"weatherApi/internal/weather/providers/tomorrowio"
	"weatherApi/internal/weather/providers/weatherapi"
)

type ProviderSet struct {
	EmailProvider        email.Provider
	TokenProvider        token.Provider
	WeatherChainProvider weather.Provider
}
type ServiceSet struct {
	SubService     subscription.Service
	WeatherService weather.Service
	EmailService   email.Service
}

func BuildProviders(cfg *config.Config, logger *log.Logger) ProviderSet {
	emailProvider := email.NewSendgrid(cfg.SendGridKey, cfg.EmailFrom)
	tokenProvider := token.NewJWT(cfg.JWTSecret)

	wrappedWeatherAPI := weather.NewWrapper(weatherapi.New(cfg.WeatherAPIKey), "WeatherAPI", logger)
	wrappedTomorrowIO := weather.NewWrapper(tomorrowio.New(cfg.TomorrowioAPIKey), "TomorrowIO", logger)

	baseWeatherAPI := weather.NewChainNode(wrappedWeatherAPI)
	baseTomorrowIO := weather.NewChainNode(wrappedTomorrowIO)

	baseWeatherAPI.SetNext(baseTomorrowIO)

	weatherChainProvider := baseWeatherAPI

	return ProviderSet{
		EmailProvider:        emailProvider,
		TokenProvider:        tokenProvider,
		WeatherChainProvider: weatherChainProvider,
	}
}

func BuildServices(db *db.DB, cfg *config.Config, p ProviderSet) ServiceSet {
	emailService := email.NewService(p.EmailProvider, cfg.BaseURL)
	tokenService := token.NewService(p.TokenProvider)
	weatherService := weather.NewService(p.WeatherChainProvider)

	repo := subscription.NewRepo(db.Gorm)
	subService := subscription.NewService(repo, emailService, weatherService, tokenService)

	return ServiceSet{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}
}
