package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"weatherApi/internal/auth"
	"weatherApi/internal/config"
	"weatherApi/internal/db"
	"weatherApi/internal/email"
	"weatherApi/internal/scheduler"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"
)

type App struct {
	Server    *http.Server
	DB        *db.DB
	Scheduler *scheduler.WeatherScheduler
}

type Services struct {
	SubService     *subscription.SubscriptionService
	WeatherService *weather.WeatherService
	EmailService   *email.EmailService
}

type ServiceBuilder struct {
	DB              *db.DB
	BaseURL         string
	EmailProvider   email.EmailProvider
	TokenProvider   auth.TokenProvider
	WeatherProvider weather.WeatherProvider
}

func (b *ServiceBuilder) BuildServices() (*Services, error) {
	emailService := email.NewEmailService(b.EmailProvider, b.BaseURL)
	tokenService := auth.NewTokenService(b.TokenProvider)
	weatherService := weather.NewWeatherService(b.WeatherProvider)

	subRepo := subscription.NewSubscriptionRepository(b.DB.Gorm)
	subService := subscription.NewSubscriptionService(subRepo, emailService, weatherService, tokenService)

	return &Services{
		SubService:     subService,
		WeatherService: weatherService,
		EmailService:   emailService,
	}, nil
}

func NewApp(cfg *config.Config) (*App, error) {
	dbInstance, err := db.NewDB(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	builder := &ServiceBuilder{
		DB:              dbInstance,
		BaseURL:         cfg.BaseURL,
		EmailProvider:   email.NewSendGridProvider(cfg.SendGridKey, cfg.EmailFrom),
		TokenProvider:   auth.NewJWTService(cfg.JWTSecret),
		WeatherProvider: weather.NewWeatherAPIProvider(cfg.WeatherAPIKey),
	}

	services, err := builder.BuildServices()
	if err != nil {
		return nil, fmt.Errorf("failed to build services: %w", err)
	}

	scheduler := scheduler.NewScheduler(services.SubService, services.WeatherService, services.EmailService)
	go scheduler.Start()

	router := SetupRoutes(services)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Server:    server,
		DB:        dbInstance,
		Scheduler: scheduler,
	}, nil
}

func (a *App) StartServer() {
	go func() {
		log.Printf("Server listening on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	a.Scheduler.Stop()
	a.DB.Close()

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Shutdown complete")
	return nil
}
