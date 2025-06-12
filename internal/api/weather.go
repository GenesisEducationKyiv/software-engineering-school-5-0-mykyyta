package api

import (
	"context"
	"net/http"
	"weatherApi/internal/weather"

	"github.com/gin-gonic/gin"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (*weather.Weather, error)
}

type WeatherHandler struct {
	service weatherService
}

func NewWeatherHandler(service weatherService) *WeatherHandler {
	return &WeatherHandler{service: service}
}

func (h *WeatherHandler) Handle(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "City is required"})
		return
	}

	data, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		// add errors
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, data)
}
