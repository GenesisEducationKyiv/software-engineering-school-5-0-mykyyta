package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	http2 "subscription/internal/adapter/email"

	weathergrpc "subscription/internal/adapter/weahtergrpc"

	"subscription/internal/adapter/gorm"
	"subscription/internal/adapter/weatherhttp"
	"subscription/internal/app/di"
	"subscription/internal/config"
	"subscription/internal/delivery"
	"subscription/internal/infra"
	"subscription/internal/subscription"
	"subscription/internal/token/jwt"

	"github.com/gin-gonic/gin"
)

type App struct {
	Server        *http.Server
	DB            *infra.Gorm
	Logger        *log.Logger
	Scheduler     *di.WeatherScheduler
	WeatherClient io.Closer
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

	if err := app.StartServer(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

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
	// DB
	db, err := infra.NewGorm(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	// Adapters
	emailClient := http2.NewClient(cfg.EmailAPIBaseURL, logger, nil)

	var weatherClient subscription.WeatherClient
	var weatherCloser io.Closer

	if cfg.UseGRPC {
		client, err := weathergrpc.NewClient(ctx, cfg.WeatherGRPCAddr)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to weather via gRPC: %w", err)
		}
		weatherClient = client
		weatherCloser = client
		logger.Println("Using gRPC weather client")
	} else {
		httpClient := weatherhttp.NewClient(cfg.WeatherHTTPAddr)
		weatherClient = httpClient
		logger.Println("Using HTTP weather client")
	}

	tokenProvider := jwt.NewJWT(cfg.JWTSecret)
	subscriptionRepo := gorm.NewRepo(db.Gorm)

	// Service
	subService := subscription.NewService(
		subscriptionRepo,
		emailClient,
		weatherClient,
		tokenProvider,
	)

	// Scheduler
	scheduler := di.NewScheduler(subService)

	// HTTP server
	router := delivery.SetupRoutes(subService, weatherClient)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Server:        server,
		DB:            db,
		Logger:        logger,
		Scheduler:     scheduler,
		WeatherClient: weatherCloser,
	}, nil
}

func (a *App) StartServer(ctx context.Context) error {
	go a.Scheduler.Start(ctx)

	go func() {
		a.Logger.Printf("Server listening on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Printf("Server error: %v", err)
		}
	}()
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Println("Shutting down application...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.Scheduler.Stop()

	if a.WeatherClient != nil {
		if err := a.WeatherClient.Close(); err != nil {
			a.Logger.Printf("Weather client close error: %v", err)
		}
	}

	a.DB.Close()

	a.Logger.Println("Shutdown complete")
	return nil
}
