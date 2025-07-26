package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription/internal/adapter/email/async"
	"subscription/internal/adapter/gorm"
	"subscription/internal/adapter/weathergrpc"
	"subscription/internal/adapter/weatherhttp"
	"subscription/internal/app/di"
	"subscription/internal/config"
	"subscription/internal/delivery"
	"subscription/internal/infra"
	"subscription/internal/infra/rabbitmq"
	"subscription/internal/subscription"
	"subscription/internal/token/jwt"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	lgCtx "subscription/pkg/logger"
)

type App struct {
	Server        *http.Server
	DB            *infra.Gorm
	Scheduler     *di.WeatherScheduler
	WeatherClient io.Closer
	RabbitMQConn  *rabbitmq.Connection
}

func Run(lg *zap.SugaredLogger) error {
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = lgCtx.With(ctx, lg)

	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	lg.Infow("subscription service started")
	defer app.DB.Close()

	if err := app.StartServer(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	lg.Infow("subscription service shutdown signal received")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx, lg); err != nil {
		return fmt.Errorf("App shutdown: %w", err)
	}

	lg.Infow("subscription service stopped")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	lg := lgCtx.From(ctx)
	// DB
	db, err := infra.NewGorm(cfg.DBUrl)
	if err != nil {
		lg.Errorw("Failed to connect to database", "err", err)
		return nil, fmt.Errorf("DB error: %w", err)
	}
	lg.Info("Connected to database")

	// Async email client
	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQUrl)
	if err != nil {
		lg.Errorw("Failed to connect to RabbitMQ", "err", err)
		return nil, fmt.Errorf("connection to RabbitMQ: %w", err)
	}
	lg.Info("Connected to RabbitMQ")
	if err := rabbitmq.VerifyExchange(rmqConn.Channel(), cfg.RabbitMQExchange, "topic"); err != nil {
		lg.Errorw("Failed to verify RabbitMQ exchange", "err", err)
		return nil, fmt.Errorf("exchange verification: %w", err)
	}
	lg.Info("Verified RabbitMQ exchange")
	rabbitPublisher := async.NewRabbitPublisher(rmqConn.Channel(), cfg.RabbitMQExchange)
	emailClient := async.NewAsyncClient(rabbitPublisher)

	// Weather client
	var weatherClient subscription.WeatherClient
	var weatherCloser io.Closer

	if cfg.UseGRPC {
		client, err := weathergrpc.NewClient(ctx, cfg.WeatherGRPCAddr)
		if err != nil {
			lg.Errorw("Failed to connect to weather via gRPC", "err", err)
			return nil, fmt.Errorf("failed to connect to weather via gRPC: %w", err)
		}
		weatherClient = client
		weatherCloser = client
		lg.Info("Using gRPC weather client")
	} else {
		httpClient := weatherhttp.NewClient(cfg.WeatherHTTPAddr)
		weatherClient = httpClient
		lg.Info("Using HTTP weather client")
	}

	// Adapters
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

	lg.Info("App is ready to serve requests")

	return &App{
		Server:        server,
		DB:            db,
		Scheduler:     scheduler,
		WeatherClient: weatherCloser,
		RabbitMQConn:  rmqConn,
	}, nil
}

func (a *App) StartServer(ctx context.Context) error {
	lg := lgCtx.From(ctx)
	go func() {
		lg.Info("Starting scheduler")
		a.Scheduler.Start(ctx)
		lg.Info("Scheduler stopped")
	}()

	go func() {
		lg.Infof("Server listening on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Errorf("Server error: %v", err)
		}
		lg.Info("HTTP server stopped")
	}()
	return nil
}

func (a *App) Shutdown(ctx context.Context, lg *zap.SugaredLogger) error {
	lg.Info("Shutting down application...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.Scheduler.Stop(ctx)
	lg.Info("Scheduler stopped")

	if err := a.RabbitMQConn.Close(); err != nil {
		lg.Errorf("Failed to close RabbitMQ connection: %v", err)
	} else {
		lg.Info("RabbitMQ connection closed")
	}

	if a.WeatherClient != nil {
		if err := a.WeatherClient.Close(); err != nil {
			lg.Errorf("Weather client close error: %v", err)
		} else {
			lg.Info("Weather client closed")
		}
	}

	a.DB.Close()
	lg.Info("Database connection closed")

	lg.Info("Shutdown complete")
	return nil
}
