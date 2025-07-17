package app

import (
	"api-gateway/internal/adapter/subscription"
	"api-gateway/internal/config"
	"api-gateway/internal/delivery"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type App struct {
	server *http.Server
	logger *log.Logger
}

func NewApp(cfg *config.Config, logger *log.Logger) *App {
	subscriptionClient := subscription.NewClient(
		cfg.SubscriptionServiceAddr,
		cfg.RequestTimeout,
	)

	mux := delivery.SetupRoutes(subscriptionClient, logger)

	server := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      mux,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	}

	return &App{
		server: server,
		logger: logger,
	}
}

func (a *App) Start() error {
	go func() {
		a.logger.Printf("API Gateway starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			a.logger.Printf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	a.logger.Println("Shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := a.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	a.logger.Println("Server exited gracefully")
	return nil
}

func Run(logger *log.Logger) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	app := NewApp(cfg, logger)
	return app.Start()
}
