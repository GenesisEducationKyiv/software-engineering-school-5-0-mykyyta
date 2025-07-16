package app

import (
	"api-gateway/internal/adapter/subscription"
	"api-gateway/internal/delivery"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"api-gateway/internal/config"
)

type App struct {
	server *http.Server
	logger *log.Logger
}

func NewApp(cfg *config.Config) (*App, error) {
	logger := log.New(os.Stdout, "[API-GATEWAY] ", log.LstdFlags)

	subscriptionClient := subscription.NewSubscriptionClient(
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
	}, nil
}

func (a *App) Start() error {
	go func() {
		a.logger.Printf("API Gateway starting on %s", a.server.Addr)
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
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
