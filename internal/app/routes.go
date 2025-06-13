package app

import (
	"net/http"

	"weatherApi/internal/handlers"
	"weatherApi/internal/subscription"
	"weatherApi/internal/weather"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(subService *subscription.SubscriptionService, weatherService *weather.WeatherService) *gin.Engine {

	router := gin.Default()

	subscribeHandler := handlers.NewSubscribeHandler(subService)
	confirmHandler := handlers.NewConfirmHandler(subService)
	unsubscribeHandler := handlers.NewUnsubscribeHandler(subService)
	weatherHandler := handlers.NewWeatherHandler(weatherService)

	api := router.Group("/api")
	{
		api.POST("/subscribe", subscribeHandler.Handle)
		api.GET("/confirm/:token", confirmHandler.Handle)
		api.GET("/unsubscribe/:token", unsubscribeHandler.Handle)
		api.GET("/weather", weatherHandler.Handle)
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

	return router
}
