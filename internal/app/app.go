package app

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"weatherApi/internal/config"
	"weatherApi/internal/db"
	"weatherApi/internal/scheduler"
)

type App struct {
	Server    *http.Server
	DB        *db.DB
	Scheduler *scheduler.WeatherScheduler
	Logger    *log.Logger
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	db, err := db.NewDB(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	providerSet := BuildProviders(cfg, logger)
	serviceSet := BuildServices(db, cfg, providerSet)

	scheduler := scheduler.New(serviceSet.SubService, serviceSet.WeatherService, serviceSet.EmailService)
	go scheduler.Start(ctx)

	router := SetupRoutes(serviceSet)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Server:    server,
		DB:        db,
		Scheduler: scheduler,
		Logger:    logger,
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
