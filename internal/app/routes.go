package app

import (
	"net/http"
	"weatherApi/internal/app/di"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"weatherApi/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(s di.Services) *gin.Engine {
	router := gin.Default()

	subscribeHandler := handlers.NewSubscribe(s.SubService)
	confirmHandler := handlers.NewConfirm(s.SubService)
	unsubscribeHandler := handlers.NewUnsubscribe(s.SubService)
	weatherHandler := handlers.NewWeatherCurrent(s.WeatherService)

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

	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

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
