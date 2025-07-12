package di

import (
	"weatherApi/monolith/internal/config"
	"weatherApi/monolith/internal/email"
	"weatherApi/monolith/internal/infra"
	"weatherApi/monolith/internal/subscription"
	"weatherApi/monolith/internal/subscription/repo"
	"weatherApi/monolith/internal/token"
	"weatherApi/monolith/internal/weather"
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
