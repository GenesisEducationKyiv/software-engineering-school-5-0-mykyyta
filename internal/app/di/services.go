package di

import (
	"weatherApi/internal/config"
	"weatherApi/internal/email"
	"weatherApi/internal/infra"
	"weatherApi/internal/subscription"
	"weatherApi/internal/token"
	"weatherApi/internal/weather"
)

type Services struct {
	SubService     subscription.Service
	WeatherService weather.Service
	EmailService   email.Service
}

func BuildServices(db *infra.Gorm, cfg *config.Config, p Providers) Services {
	emailService := email.NewService(p.EmailProvider, cfg.BaseURL)
	tokenService := token.NewService(p.TokenProvider)
	weatherService := weather.NewService(p.WeatherChainProvider)

	repo := subscription.NewRepo(db.Gorm)
	subService := subscription.NewService(repo, emailService, weatherService, tokenService)

	return Services{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}
}
