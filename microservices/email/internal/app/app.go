package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"email/internal/app/di"
	"email/internal/delivery/consumer"
	infra "email/internal/infra/redis"

	"github.com/pkg/errors"

	"email/internal/adapter/sendgrid"
	"email/internal/adapter/template"
	"email/internal/config"
	"email/internal/delivery"
	"email/internal/email"
)

type App struct {
	Server        *http.Server
	Logger        *log.Logger
	QueueConsumer *consumer.Consumer
	ShutdownFunc  func() error
}

func Run(logger *log.Logger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}

	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	logger.Println("Shutdown signal received")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	logger.Println("Server exited gracefully")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	templateStore, err := template.Load("template")
	if err != nil {
		logger.Printf("Loading templates: %v", err)
		return nil, err
	}

	emailProvider := sendgrid.New(cfg.SendGridKey, cfg.EmailFrom)
	emailService := email.NewService(emailProvider, templateStore)

	redisClient, err := infra.NewRedisClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	queueModule, err := di.NewQueueModule(ctx, cfg, emailService, redisClient, logger)
	if err != nil {
		logger.Printf("Failed to init queue module: %v", err)
		return nil, err
	}

	handler := delivery.NewEmailHandler(emailService, logger)

	mux := http.NewServeMux()
	delivery.RegisterRoutes(mux, handler)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	return &App{
		Server:        server,
		Logger:        logger,
		QueueConsumer: queueModule.Consumer,
		ShutdownFunc:  queueModule.ShutdownFunc,
	}, nil
}

func (a *App) Start(ctx context.Context) error {
	go func() {
		a.Logger.Printf("Email service running at %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.Logger.Printf("Server error: %v", err)
		}
	}()

	go func() {
		a.Logger.Println("Starting async consumer...")
		if err := a.QueueConsumer.Start(ctx); err != nil {
			a.Logger.Printf("Consumer error: %v", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	a.Logger.Println("Shutting down email service...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.Logger.Println("Shutdown complete")
	return nil
}
