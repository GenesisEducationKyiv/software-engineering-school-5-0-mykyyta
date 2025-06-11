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

	"github.com/gin-gonic/gin"

	"weatherApi/config"
	"weatherApi/internal/api"
	"weatherApi/internal/db"
	"weatherApi/internal/scheduler"
)

func Run() error {
	// Set GIN mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	log.Printf("🚀 Starting in %s mode\n", gin.Mode())

	// Load configuration and connect to DB
	config.LoadConfig()
	db.ConnectDefaultDB()
	defer db.CloseDB()

	// Inject DB into other modules
	api.SetDB(db.DB)
	scheduler.SetDB(db.DB)

	// Start background scheduler
	go scheduler.StartWeatherScheduler()

	// Set up Gin router
	r := gin.Default()
	api.RegisterRoutes(r)

	// Create HTTP server
	srv := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: r,
	}

	// Start server in background
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for termination signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, cleaning up...")

	// Gracefully shutdown HTTP server
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	log.Println("✅ Server exited gracefully")
	return nil
}
