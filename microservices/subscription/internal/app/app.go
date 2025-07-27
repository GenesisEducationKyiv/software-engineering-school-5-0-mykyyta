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

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type App struct {
	Server        *http.Server
	DB            *infra.Gorm
	Scheduler     *di.WeatherScheduler
	WeatherClient io.Closer
	RabbitMQConn  *rabbitmq.Connection
}

func Run(logger *zap.SugaredLogger) error {
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = loggerPkg.With(ctx, logger)

	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	logger.Infow("subscription service started")
	defer app.DB.Close(ctx)

	if err := app.StartServer(ctx); err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Infow("subscription service shutdown signal received")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("App shutdown: %w", err)
	}

	logger.Infow("subscription service stopped")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	logger := loggerPkg.From(ctx)
	// DB
	db, err := infra.NewGorm(cfg.DBUrl)
	if err != nil {
		logger.Errorw("Failed to connect to database", "err", err)
		return nil, fmt.Errorf("DB error: %w", err)
	}
	logger.Info("Connected to database")

	// Async email client
	rmqConn, err := rabbitmq.NewConnection(cfg.RabbitMQUrl)
	if err != nil {
		logger.Errorw("Failed to connect to RabbitMQ", "err", err)
		return nil, fmt.Errorf("connection to RabbitMQ: %w", err)
	}
	logger.Info("Connected to RabbitMQ")
	if err := rabbitmq.VerifyExchange(rmqConn.Channel(), cfg.RabbitMQExchange, "topic"); err != nil {
		logger.Errorw("Failed to verify RabbitMQ exchange", "err", err)
		return nil, fmt.Errorf("exchange verification: %w", err)
	}
	logger.Info("Verified RabbitMQ exchange")
	rabbitPublisher := async.NewRabbitPublisher(rmqConn.Channel(), cfg.RabbitMQExchange)
	emailClient := async.NewAsyncClient(rabbitPublisher)

	// Weather client
	var weatherClient subscription.WeatherClient
	var weatherCloser io.Closer

	if cfg.UseGRPC {
		client, err := weathergrpc.NewClient(ctx, cfg.WeatherGRPCAddr)
		if err != nil {
			logger.Errorw("Failed to connect to weather via gRPC", "err", err)
			return nil, fmt.Errorf("failed to connect to weather via gRPC: %w", err)
		}
		weatherClient = client
		weatherCloser = client
		logger.Info("Using gRPC weather client")
	} else {
		httpClient := weatherhttp.NewClient(cfg.WeatherHTTPAddr)
		weatherClient = httpClient
		logger.Info("Using HTTP weather client")
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
	router := delivery.SetupRoutes(subService, weatherClient, logger)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	logger.Info("App is ready to serve requests")

	return &App{
		Server:        server,
		DB:            db,
		Scheduler:     scheduler,
		WeatherClient: weatherCloser,
		RabbitMQConn:  rmqConn,
	}, nil
}

func (a *App) StartServer(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	go func() {
		logger.Info("Starting scheduler")
		a.Scheduler.Start(ctx)
		logger.Info("Scheduler stopped")
	}()

	go func() {
		logger.Infof("Server listening on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Server error: %v", err)
		}
		logger.Info("HTTP server stopped")
	}()
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	logger.Info("Shutting down application...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.Scheduler.Stop(ctx)
	logger.Info("Scheduler stopped")

	if err := a.RabbitMQConn.Close(); err != nil {
		logger.Errorf("Failed to close RabbitMQ connection: %v", err)
	} else {
		logger.Info("RabbitMQ connection closed")
	}

	if a.WeatherClient != nil {
		if err := a.WeatherClient.Close(); err != nil {
			logger.Errorf("Weather client close error: %v", err)
		} else {
			logger.Info("Weather client closed")
		}
	}

	a.DB.Close(ctx)
	logger.Info("Database connection closed")

	logger.Info("Shutdown complete")
	return nil
}
