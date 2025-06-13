package app

import (
	"net/http"
	"weatherApi/internal/handler"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(subService *subscription.SubscriptionService, weatherService *weather.WeatherService) *gin.Engine {

	router := gin.Default()

	// Створення handler-ів
	subscribeHandler := handler.NewSubscribeHandler(subService)
	confirmHandler := handler.NewConfirmHandler(subService)
	unsubscribeHandler := handler.NewUnsubscribeHandler(subService)
	weatherHandler := handler.NewWeatherHandler(weatherService)

	// API група
	api := router.Group("/api")
	{
		api.POST("/subscribe", subscribeHandler.Handle)
		api.GET("/confirm/:token", confirmHandler.Handle)
		api.GET("/unsubscribe/:token", unsubscribeHandler.Handle)
		api.GET("/weather", weatherHandler.Handle)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// HTML-шаблони
	if gin.Mode() != gin.TestMode {
		router.LoadHTMLGlob("templates/*.html")
	}

	router.GET("/subscribe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "subscribe.html", nil)
	})
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/subscribe")
	})

	return router
}
