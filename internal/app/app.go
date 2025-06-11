package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"weatherApi/config"
	"weatherApi/internal/api"
	"weatherApi/internal/db"
	"weatherApi/pkg/scheduler"
)

func Run() error {
	// Set GIN mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	log.Printf("🚀 Starting in %s mode\n", gin.Mode())

	// Load configuration
	config.LoadConfig()

	// Connect to DB
	db.ConnectDefaultDB()
	dbInstance := db.DB

	// Inject DB
	api.SetDB(dbInstance)
	scheduler.SetDB(dbInstance)

	// Graceful shutdown context
	_, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background scheduler
	go scheduler.StartWeatherScheduler()

	// Handle OS signals
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigs
		log.Println("🔌 Shutting down gracefully...")
		cancel()
		time.Sleep(2 * time.Second)
		os.Exit(0)
	}()

	// Init and run server
	r := gin.Default()
	api.RegisterRoutes(r)
	return r.Run(":" + config.C.Port)
}
