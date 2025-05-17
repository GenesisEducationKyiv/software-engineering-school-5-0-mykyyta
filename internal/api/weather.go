package api

import (
	"net/http"

	"weatherApi/pkg/weatherapi"

	"github.com/gin-gonic/gin"
)

// getWeatherHandler retrieves current weather for a given city.
// This endpoint is intended for real-time weather preview (e.g., before subscribing).
// It requires a "city" query parameter and responds with weather data in JSON.
func getWeatherHandler(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		// Client must specify a city name in query parameters
		c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
		return
	}

	// Fetch weather using external API and return appropriate status code
	weather, statusCode, err := weatherapi.FetchWithStatus(city)
	if err != nil {
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, weather)
}
