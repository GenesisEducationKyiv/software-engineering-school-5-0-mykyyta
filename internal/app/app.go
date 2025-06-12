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

	"weatherApi/internal/scheduler"

	"github.com/gin-gonic/gin"

	"weatherApi/config"
	"weatherApi/internal/api"
	"weatherApi/internal/db"
	"weatherApi/internal/email"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"
)

func Run() error {
	// GIN mode
	mode := os.Getenv("GIN_MODE")
	if mode == "" {
		mode = gin.DebugMode
	}
	gin.SetMode(mode)
	log.Printf("Starting in %s mode\n", gin.Mode())

	// Load config
	config.LoadConfig()

	// Connect DB
	dbInstance := db.ConnectDefaultDB()
	defer db.CloseDB(dbInstance)

	// === ІНІЦІАЛІЗАЦІЯ ЗАЛЕЖНОСТЕЙ ===

	// Weather service
	weatherProvider := weather.NewWeatherAPIProvider(config.C.WeatherAPIKey)
	weatherService := weather.NewService(weatherProvider)

	// Subscription service
	repo := subscription.NewSubscriptionRepository(dbInstance)

	emailProvider := email.NewSendGridProvider(config.C.EmailFrom, config.C.SendGridKey)
	emailService := email.NewEmailService(emailProvider)
	subService := subscription.NewSubscriptionService(repo, emailService, weatherService)

	// Scheduler
	sched := scheduler.NewScheduler(subService, weatherService, emailService)
	go sched.Start()

	// Handlers
	subscribeHandler := api.NewSubscribeHandler(subService)
	confirmHandler := api.NewConfirmHandler(subService)
	unsubscribeHandler := api.NewUnsubscribeHandler(subService)
	weatherHandler := api.NewWeatherHandler(weatherService)

	// === РОУТИНГ ===
	router := gin.Default()

	apiGroup := router.Group("/api")
	{
		apiGroup.POST("/subscribe", subscribeHandler.Handle)
		apiGroup.GET("/confirm/:token", confirmHandler.Handle)
		apiGroup.GET("/unsubscribe/:token", unsubscribeHandler.Handle)
		apiGroup.GET("/weather", weatherHandler.Handle)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if gin.Mode() != gin.TestMode {
		router.LoadHTMLGlob("templates/*.html")
	}

	router.GET("/subscribe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "subscribe.html", nil)
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/subscribe")
	})

	// server run
	srv := &http.Server{
		Addr:    ":" + config.C.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutdown signal received, cleaning up...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("HTTP server shutdown: %w", err)
	}

	log.Println("Server exited gracefully")
	return nil
}
