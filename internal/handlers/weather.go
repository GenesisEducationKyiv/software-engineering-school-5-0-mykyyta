package handlers

import (
	"context"
	"errors"
	"net/http"

	weather2 "weatherApi/internal/weather"

	"weatherApi/internal/weather"

	"github.com/gin-gonic/gin"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (weather.Weather, error)
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
		SendError(c, http.StatusBadRequest, "City is required")
		return
	}

	data, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		switch {
		case errors.Is(err, weather2.ErrCityNotFound):
			SendError(c, http.StatusBadRequest, "City not found")
		default:
			SendError(c, http.StatusInternalServerError, "Failed to fetch weather data")
		}
		return
	}

	c.JSON(http.StatusOK, data)
}
