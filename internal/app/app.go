package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"weatherApi/internal/weather/cache"

	"github.com/redis/go-redis/v9"

	"weatherApi/internal/config"
	"weatherApi/internal/scheduler"
)

type App struct {
	Server    *http.Server
	DB        *DB
	Redis     *redis.Client
	Scheduler *scheduler.WeatherScheduler
	Logger    *log.Logger
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	db, err := NewDB(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	redisClient, err := NewRedisClient(ctx, cfg)
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("redis error: %w", err)
	}

	metrics := cache.NewMetrics()
	metrics.Register()

	httpClient := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:    100,
			IdleConnTimeout: 90 * time.Second,
		},
	}

	providerDeps := ProviderDeps{
		cfg:         cfg,
		logger:      logger,
		redisClient: redisClient,
		httpClient:  httpClient,
		metrics:     metrics,
	}

	providerSet := BuildProviders(providerDeps)
	serviceSet := BuildServices(db, cfg, providerSet)

	sr := scheduler.New(serviceSet.SubService, serviceSet.WeatherService, serviceSet.EmailService)
	go sr.Start(ctx)

	router := SetupRoutes(serviceSet)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Server:    server,
		DB:        db,
		Redis:     redisClient,
		Scheduler: sr,
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

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Printf("Redis close error: %v", err)
		}
	}

	a.DB.Close()

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Shutdown complete")
	return nil
}
