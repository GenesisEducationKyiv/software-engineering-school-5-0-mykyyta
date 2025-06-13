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
	log.Printf("GIN running in %s mode", gin.Mode())

	app, err := NewApp(cfg)
	if err != nil {
		return fmt.Errorf("failed to build app: %w", err)
	}
	defer app.DB.Close()

	app.StartServer()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, cleaning up...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	log.Println("Server exited gracefully")
	return nil
}
