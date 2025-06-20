package app

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"weatherApi/internal/config"
)

func Run() error {
	cfg := config.LoadConfig()
	gin.SetMode(cfg.GinMode)

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("cannot open log file: %w", err)
	}
	logger := log.New(logFile, "", log.LstdFlags)

	logger.Printf("GIN running in %s mode", gin.Mode())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app, err := NewApp(ctx, cfg, logger)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	defer app.DB.Close()

	app.StartServer()

	// Graceful shutdown
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

	if err := logFile.Close(); err != nil {
		return fmt.Errorf("failed to close log file: %w", err)
	}
	return nil
}
