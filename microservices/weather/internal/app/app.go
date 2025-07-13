package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weather/internal/adapter/cache"

	"weather/internal/delivery"
	"weather/internal/delivery/handler"

	"github.com/redis/go-redis/v9"

	"weather/internal/app/di"
	"weather/internal/config"
	"weather/internal/infra"
	"weather/internal/service"
)

type App struct {
	Server *http.Server
	Redis  *redis.Client
	Logger *log.Logger
}

func Run(logger *log.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("build app: %w", err)
	}
	defer func() {
		if err := app.Redis.Close(); err != nil {
			logger.Printf("close Redis: %v", err)
		}
	}()

	app.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Println("Shutdown signal received")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	return app.Shutdown(shutdownCtx)
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	redisClient, err := infra.NewRedisClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	metrics := cache.NewMetrics()
	metrics.Register()

	httpClient := &http.Client{Timeout: 5 * time.Second}

	weatherProvider := di.BuildProviders(di.ProviderDeps{
		Cfg:         cfg,
		Logger:      logger,
		RedisClient: redisClient,
		HttpClient:  httpClient,
		Metrics:     metrics,
	})

	weatherService := service.NewService(weatherProvider)

	mux := http.NewServeMux()
	weatherHandler := handler.NewWeatherHandler(weatherService)
	delivery.RegisterRoutes(mux, weatherHandler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	return &App{
		Server: server,
		Redis:  redisClient,
		Logger: logger,
	}, nil
}

func (a *App) Start() {
	go func() {
		a.Logger.Printf("Weather service running at %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Fatalf("server error: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	if a.Redis != nil {
		if err := a.Redis.Close(); err != nil {
			a.Logger.Printf("Redis close error: %v", err)
		}
	}

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Shutdown complete")
	return nil
}
