package app

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"subscription/internal/adapter/email"
	"subscription/internal/adapter/gorm"
	"subscription/internal/adapter/weather"
	"subscription/internal/app/di"
	"subscription/internal/config"
	"subscription/internal/delivery"
	"subscription/internal/infra"
	"subscription/internal/service"
	"subscription/internal/token/jwt"

	"github.com/gin-gonic/gin"
)

type App struct {
	Server    *http.Server
	DB        *infra.Gorm
	Logger    *log.Logger
	Scheduler *di.WeatherScheduler
}

func Run(logger *log.Logger) error {
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	defer app.DB.Close()

	app.StartServer()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Println("Shutdown signal received, cleaning up...")

	cancel()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("App shutdown: %w", err)
	}

	logger.Println("Server exited gracefully")
	return nil
}

func NewApp(ctx context.Context, cfg *config.Config, logger *log.Logger) (*App, error) {
	// DB
	db, err := infra.NewGorm(cfg.DBUrl)
	if err != nil {
		return nil, fmt.Errorf("DB error: %w", err)
	}

	// Adapters
	emailClient := email.NewClient(cfg.EmailAPIBaseURL, logger, nil)
	weatherClient := weather.NewClient(cfg.WeatherAPIBaseURL)
	tokenProvider := jwt.NewJWT(cfg.JWTSecret)
	subscriptionRepo := gorm.NewRepo(db.Gorm)

	// Service
	subService := service.NewService(
		subscriptionRepo,
		emailClient,
		weatherClient,
		tokenProvider,
	)

	// Scheduler
	scheduler := di.NewScheduler(subService)
	go scheduler.Start(ctx)

	// HTTP server
	router := delivery.SetupRoutes(subService, weatherClient)
	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	return &App{
		Server:    server,
		DB:        db,
		Logger:    logger,
		Scheduler: scheduler,
	}, nil
}

func (a *App) StartServer() {
	go func() {
		log.Printf("Server listening on %s", a.Server.Addr)
		if err := a.Server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Server error: %v", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) error {
	log.Println("Shutting down application...")

	a.Scheduler.Stop()
	a.DB.Close()

	if err := a.Server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown error: %w", err)
	}

	log.Println("Shutdown complete")
	return nil
}
