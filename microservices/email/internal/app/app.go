package app

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"

	"email/internal/adapter/sendgrid"

	"email/internal/app/di"
	"email/internal/delivery/consumer"
	infra "email/internal/infra/redis"

	"github.com/pkg/errors"

	"email/internal/adapter/template"
	"email/internal/config"
	"email/internal/delivery"
	"email/internal/email"
	"email/pkg/logger"
)

type App struct {
	Server        *http.Server
	QueueConsumer *consumer.Consumer
	ShutdownFunc  func() error
}

func Run(lg *zap.SugaredLogger) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ctx = logger.With(ctx, lg)

	cfg := config.LoadConfig()

	app, err := NewApp(ctx, cfg)
	if err != nil {
		return fmt.Errorf("creating application: %w", err)
	}

	if err := app.Start(ctx); err != nil {
		return fmt.Errorf("starting server: %w", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	lg.Infow("Shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown error: %w", err)
	}

	lg.Infow("Server exited gracefully")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config) (*App, error) {
	templateStore, err := template.Load("template")
	if err != nil {
		logger.From(ctx).Errorw("Loading templates", "err", err)
		return nil, err
	}

	emailProvider := sendgrid.New(cfg.SendGridKey, cfg.EmailFrom)
	emailService := email.NewService(emailProvider, templateStore)

	redisClient, err := infra.NewRedisClient(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("redis error: %w", err)
	}

	queueModule, err := di.NewQueueModule(ctx, cfg, emailService, redisClient)
	if err != nil {
		logger.From(ctx).Errorw("Failed to init queue module", "err", err)
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
	logg := logger.From(ctx)
	go func() {
		logg.Infow("Email service running", "addr", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logg.Errorw("Server error", "err", err)
		}
	}()

	go func() {
		logg.Infow("Starting async consumer...")
		if err := a.QueueConsumer.Start(ctx); err != nil {
			logg.Errorw("Consumer error", "err", err)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	logg := logger.From(ctx)
	logg.Infow("Shutting down email service...")

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	logg.Infow("Shutdown complete")
	return nil
}
