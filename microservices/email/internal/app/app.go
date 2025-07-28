package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email/internal/adapter/sendgrid"

	"email/internal/app/di"
	"email/internal/delivery/consumer"
	infra "email/internal/infra/redis"

	"github.com/pkg/errors"

	"email/internal/adapter/template"
	"email/internal/config"
	"email/internal/delivery"
	"email/internal/email"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type App struct {
	Server        *http.Server
	QueueConsumer *consumer.Consumer
	ShutdownFunc  func() error
}

func Run(logger *loggerPkg.Logger) error {
	logger.Info("Starting email service application")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = loggerPkg.With(ctx, logger)

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg)
	if err != nil {
		logger.Error("Failed to create application", "err", err)
		return fmt.Errorf("creating application: %w", err)
	}

	if err := app.Start(ctx); err != nil {
		logger.Error("Failed to start application", "err", err)
		return fmt.Errorf("starting server: %w", err)
	}

	logger.Info("Email service started successfully", "http_port", cfg.Port)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Info("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		logger.Error("Shutdown failed", "err", err)
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Info("Server exited gracefully")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	templateStore, err := template.Load("template")
	if err != nil {
		loggerPkg.From(ctx).Error("Failed to load email templates", "template_dir", "template", "err", err)
		return nil, err
	}

	emailProvider := sendgrid.New(cfg.SendGridKey, cfg.EmailFrom)
	emailService := email.NewService(emailProvider, templateStore)

	redisClient, err := infra.NewRedisClient(ctx, cfg)
	if err != nil {
		loggerPkg.From(ctx).Error("Failed to connect to Redis", "redis_url", cfg.RedisURL, "err", err)
		return nil, fmt.Errorf("redis error: %w", err)
	}

	queueModule, err := di.NewQueueModule(ctx, cfg, emailService, redisClient)
	if err != nil {
		loggerPkg.From(ctx).Error("Failed to init queue module", "err", err)
		return nil, err
	}

	handler := delivery.NewEmailHandler(emailService)

	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	return &App{
		Server:        server,
		QueueConsumer: queueModule.Consumer,
		ShutdownFunc:  queueModule.ShutdownFunc,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	go func() {
		logger.Info("Email service running", "addr", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("Server error", "err", err)
		}
	}()

	go func() {
		logger.Info("Starting async consumer...")
		if err := a.QueueConsumer.Start(ctx); err != nil {
			logger.Error("Consumer error", "err", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logger := loggerPkg.From(ctx)
	logger.Info("Shutting down email service...")

	if err := a.Server.Shutdown(ctx); err != nil {
		logger.Error("HTTP server shutdown failed", "err", err)
		return fmt.Errorf("server shutdown error: %w", err)
	}

	if a.ShutdownFunc != nil {
		if err := a.ShutdownFunc(); err != nil {
			logger.Error("Resource cleanup failed", "err", err)
		}
	}

	logger.Info("Email service shutdown completed")
	return nil
}
