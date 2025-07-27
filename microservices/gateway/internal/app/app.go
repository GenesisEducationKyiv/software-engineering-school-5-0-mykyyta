package app

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"gateway/internal/adapter/subscription"
	"gateway/internal/config"
	"gateway/internal/delivery"
	"gateway/internal/service"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"go.uber.org/zap"
)

type App struct {
	server *http.Server
}

func NewApp(cfg *config.Config, ctx context.Context) *App {
	subscriptionClient := subscription.NewClient(
		cfg.SubscriptionServiceAddr,
		cfg.RequestTimeout,
	)
	validator := service.NewSecurityValidator()
	gatewayService := service.NewService(subscriptionClient, validator)
	responseWriter := delivery.NewResponseWriter()
	handler := delivery.NewSubscriptionHandler(gatewayService, responseWriter)

	mux := delivery.SetupRoutes(handler, cfg, ctx)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &App{
		server: server,
	}
}

func (a *App) Start(ctx context.Context) error {
	logger := loggerPkg.From(ctx)

	logger.Infow("API Gateway starting", "addr", a.server.Addr)

	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorw("Server error", "err", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Infow("Shutting down API Gateway")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(shutdownCtx); err != nil {
		logger.Errorw("Server shutdown failed", "err", err)
		return err
	}

	logger.Infow("API Gateway shutdown completed")
	return nil
}

func Run(logger *zap.SugaredLogger) error {
	cfg, err := config.Load()
	if err != nil {
		logger.Errorw("Failed to load configuration", "err", err)
		return err
	}

	ctx := loggerPkg.With(context.Background(), logger)

	app := NewApp(cfg, ctx)
	return app.Start(ctx)
}
