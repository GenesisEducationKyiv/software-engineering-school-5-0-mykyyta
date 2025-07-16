package delivery

import (
	"net/http"
	"subscription/internal/domain"

	"golang.org/x/net/context"

	"subscription/internal/subscription"

	subscription2 "subscription/internal/delivery/handlers/subscription"
	handlers2 "subscription/internal/delivery/handlers/weather"

	"github.com/gin-gonic/gin"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
}

func SetupRoutes(subService subscription.Service, weatherClient weatherService) *gin.Engine {
	router := gin.Default()

	subscribeHandler := subscription2.NewSubscribe(subService)
	confirmHandler := subscription2.NewConfirm(subService)
	unsubscribeHandler := subscription2.NewUnsubscribe(subService)
	weatherHandler := handlers2.NewWeatherCurrent(weatherClient)

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
