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

	di2 "monolith/internal/app/di"
	"monolith/internal/config"
	"monolith/internal/delivery"
	infra2 "monolith/internal/infra"
	"monolith/internal/weather/cache"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type App struct {
	Server    *http.Server
	DB        *infra2.Gorm
	Redis     *redis.Client
	Scheduler *di2.WeatherScheduler
	Logger    *log.Logger
}

func Run(logger *log.Logger) error {
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	defer app.DB.Close()

	app.StartServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutdown signal received, cleaning up...")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("App shutdown: %w", err)
	}

	logger.Println("Server exited gracefully")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	db, err := infra2.NewGorm(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	redisClient, err := infra2.NewRedisClient(ctx, cfg)
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

	providerDeps := di2.ProviderDeps{
		Cfg:         cfg,
		Logger:      logger,
		RedisClient: redisClient,
		HttpClient:  httpClient,
		Metrics:     metrics,
	}

	providerSet := di2.BuildProviders(providerDeps)
	serviceSet := di2.BuildServices(db, cfg, providerSet)

	sr := di2.NewScheduler(serviceSet.SubService)
	go sr.Start(ctx)

	router := delivery.SetupRoutes(serviceSet)

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
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
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
