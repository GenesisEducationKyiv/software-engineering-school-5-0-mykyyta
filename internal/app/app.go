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

func NewApp(cfg *config.Config) (*App, error) {
	dbInstance, err := db.NewDB(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	weatherProvider := weather.NewWeatherAPIProvider(cfg.WeatherAPIKey)
	weatherService := weather.NewService(weatherProvider)

	emailProvider := email.NewSendGridProvider(cfg.EmailFrom, cfg.SendGridKey)
	emailService := email.NewEmailService(emailProvider, cfg.BaseURL)

	tokenService := auth.NewJWTService(cfg.JWTSecret)
	subRepo := subscription.NewSubscriptionRepository(dbInstance.Gorm)
	subService := subscription.NewSubscriptionService(subRepo, emailService, weatherService, tokenService)

	scheduler := scheduler.NewScheduler(subService, weatherService, emailService)
	go scheduler.Start()

	router := SetupRoutes(subService, weatherService)

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

	a.Scheduler.Stop() // якщо маєш Stop(), або просто залиш пустим
	a.DB.Close()

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Shutdown complete")
	return nil
}
