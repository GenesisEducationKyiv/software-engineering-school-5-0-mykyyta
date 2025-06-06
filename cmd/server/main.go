package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"weatherApi/config"
	"weatherApi/internal/api"
	"weatherApi/internal/db"
	"weatherApi/pkg/scheduler"

	"github.com/gin-gonic/gin"
)

func main() {
	// ─── GIN Mode Setup ─────────────────────────────────────
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode // ← default DEBUG
	}
	gin.SetMode(mode)
	log.Printf("🚀 Starting in %s mode\n", gin.Mode())

	// Load application configuration (from environment or .env file)
	config.LoadConfig()

	// Initialize and connect to the database
	db.ConnectDefaultDB()
	dbInstance := db.DB

	// Inject DB instance into API and scheduler layers
	api.SetDB(dbInstance)
	scheduler.SetDB(dbInstance)

	// Set up graceful shutdown context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background weather update scheduler in a separate goroutine
	go scheduler.StartWeatherScheduler()

	// Listen for termination signals (e.g., Ctrl+C, SIGTERM from Docker/K8s)
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("Shutting down gracefully...")
		cancel()
		time.Sleep(2 * time.Second) // Give time for background tasks to finish
		os.Exit(0)
	}()

	// Initialize Gin HTTP server and register all routes
	r := gin.Default()
	api.RegisterRoutes(r)

	// Start HTTP server on the configured port
	if err := r.Run(":" + config.C.Port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
