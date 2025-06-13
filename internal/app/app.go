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
	"weatherApi/internal/auth"
	"weatherApi/internal/config"

	"weatherApi/internal/scheduler"

	"github.com/gin-gonic/gin"

	"weatherApi/internal/db"
	"weatherApi/internal/email"
	"weatherApi/internal/handler"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"
)

func Run() error {
	cfg := config.LoadConfig()

	gin.SetMode(cfg.GinMode)
	log.Printf("GIN running in %s mode", gin.Mode())

	dbInstance, err := db.NewDB(cfg.DBUrl)
	if err != nil {
		log.Fatalf("failed to init DB: %v", err)
	}
	defer dbInstance.Close()

	weatherProvider := weather.NewWeatherAPIProvider(cfg.WeatherAPIKey)
	weatherService := weather.NewService(weatherProvider)
	emailProvider := email.NewSendGridProvider(cfg.EmailFrom, cfg.SendGridKey)
	emailService := email.NewEmailService(emailProvider, cfg.BaseURL)
	tokenService := auth.NewJWTService(cfg.JWTSecret)
	subRepo := subscription.NewSubscriptionRepository(dbInstance.Gorm)
	subService := subscription.NewSubscriptionService(subRepo, emailService, weatherService, tokenService)

	sched := scheduler.NewScheduler(subService, weatherService, emailService)
	go sched.Start()

	subscribeHandler := handler.NewSubscribeHandler(subService)
	confirmHandler := handler.NewConfirmHandler(subService)
	unsubscribeHandler := handler.NewUnsubscribeHandler(subService)
	weatherHandler := handler.NewWeatherHandler(weatherService)

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
		Addr:    ":" + cfg.Port,
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
