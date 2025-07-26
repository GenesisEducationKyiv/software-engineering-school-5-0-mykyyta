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
	"gateway/internal/middleware"
	"gateway/internal/service"
	loggerCtx "gateway/pkg/logger"

	"go.uber.org/zap"
)

type App struct {
	server *http.Server
}

func NewApp(cfg *config.Config, lg *zap.SugaredLogger) *App {
	subscriptionClient := subscription.NewClient(
		cfg.SubscriptionServiceAddr,
		cfg.RequestTimeout,
	)
	validator := service.NewSecurityValidator()
	gatewayService := service.NewService(subscriptionClient, validator)
	responseWriter := delivery.NewResponseWriter()
	handler := delivery.NewSubscriptionHandler(gatewayService, responseWriter)

	mux := delivery.SetupRoutes(handler, cfg)
	mux = middleware.WithLogger(lg)(mux)

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
	lg := loggerCtx.From(ctx)
	go func() {
		lg.Infof("API Gateway starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			lg.Errorf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	lg.Info("Shutting down API Gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.server.Shutdown(shutdownCtx)
}

func Run(lg *zap.SugaredLogger) error {
	cfg, err := config.Load()
	if err != nil {
		lg.Errorf("failed to load config: %v", err)
		return err
	}
	app := NewApp(cfg, lg)
	ctx := loggerCtx.With(context.Background(), lg)
	return app.Start(ctx)
}
