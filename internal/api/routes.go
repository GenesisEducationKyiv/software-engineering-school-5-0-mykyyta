package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterRoutes wires all API and UI routes.
// Only development-safe routes should be exposed in production builds.
func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api")
	{
		api.POST("/subscribe", subscribeHandler)
		api.GET("/confirm/:token", confirmHandler)
		api.GET("/unsubscribe/:token", unsubscribeHandler)
		api.GET("/weather", getWeatherHandler)
	}

	// Register only in non-production mode
	if gin.Mode() != gin.ReleaseMode {
		r.GET("/subscriptions", listSubscriptionsHandler)
	}

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	if gin.Mode() != gin.TestMode {
		r.LoadHTMLGlob("templates/*.html")
		r.Static("/static", "./static")
	}

	r.GET("/subscribe", func(c *gin.Context) {
		c.HTML(http.StatusOK, "subscribe.html", nil)
	})

	r.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/subscribe")
	})
}
