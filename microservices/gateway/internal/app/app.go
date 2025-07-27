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

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"go.uber.org/zap"
)

type App struct {
	server *http.Server
}

func NewApp(cfg *config.Config, logger *zap.SugaredLogger) *App {
	subscriptionClient := subscription.NewClient(
		cfg.SubscriptionServiceAddr,
		cfg.RequestTimeout,
	)
	validator := service.NewSecurityValidator()
	gatewayService := service.NewService(subscriptionClient, validator)
	responseWriter := delivery.NewResponseWriter()
	handler := delivery.NewSubscriptionHandler(gatewayService, responseWriter)

	mux := delivery.SetupRoutes(handler, cfg)
	mux = middleware.WithLogger(logger)(mux)

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
	go func() {
		logger.Infof("API Gateway starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Errorf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down API Gateway...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.server.Shutdown(shutdownCtx)
}

func Run(logger *zap.SugaredLogger) error {
	cfg, err := config.Load()
	if err != nil {
		logger.Errorf("failed to load config: %v", err)
		return err
	}
	app := NewApp(cfg, logger)
	ctx := loggerPkg.With(context.Background(), logger)
	return app.Start(ctx)
}
