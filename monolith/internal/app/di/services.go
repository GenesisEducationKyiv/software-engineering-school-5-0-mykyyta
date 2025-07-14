package di

import (
	"monolith/internal/config"
	"monolith/internal/email"
	"monolith/internal/infra"
	"monolith/internal/subscription"
	"monolith/internal/subscription/repo"
	"monolith/internal/token"
	"monolith/internal/weather"
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

	repo := repo.NewRepo(db.Gorm)
	subService := subscription.NewService(repo, emailService, weatherService, tokenService)

	return Services{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}
}
