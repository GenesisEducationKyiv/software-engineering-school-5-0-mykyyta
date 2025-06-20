package app

import (
	"net/http"

	"weatherApi/internal/handlers"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(s *ServiceSet) *gin.Engine {
	router := gin.Default()

	subscribeHandler := handlers.NewSubscribeHandler(s.SubService)
	confirmHandler := handlers.NewConfirmHandler(s.SubService)
	unsubscribeHandler := handlers.NewUnsubscribeHandler(s.SubService)
	weatherHandler := handlers.NewWeatherHandler(s.WeatherService)

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
