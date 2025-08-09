package delivery

import (
	"net/http"

	"monolith/internal/app/di"
	subscription2 "monolith/internal/delivery/handlers/subscription"
	handlers2 "monolith/internal/delivery/handlers/weather"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func SetupRoutes(s di.Services) *gin.Engine {
	router := gin.Default()

	subscribeHandler := subscription2.NewSubscribe(s.SubService)
	confirmHandler := subscription2.NewConfirm(s.SubService)
	unsubscribeHandler := subscription2.NewUnsubscribe(s.SubService)
	weatherHandler := handlers2.NewWeatherCurrent(s.WeatherService)

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
