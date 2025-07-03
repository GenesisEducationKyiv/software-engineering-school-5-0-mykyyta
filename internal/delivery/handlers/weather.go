package handlers

import (
	"context"
	"errors"
	"net/http"

	"weatherApi/internal/weather"

	"github.com/gin-gonic/gin"
)

type weatherCurrent interface {
	GetWeather(ctx context.Context, city string) (weather.Report, error)
}

type WeatherCurrent struct {
	service weatherCurrent
}

func NewWeatherCurrent(service weatherCurrent) *WeatherCurrent {
	return &WeatherCurrent{service: service}
}

func (h WeatherCurrent) Handle(c *gin.Context) {
	city := c.Query("city")
	if city == "" {
		SendError(c, http.StatusBadRequest, "City is required")
		return
	}

	data, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		switch {
		case errors.Is(err, weather.ErrCityNotFound):
			SendError(c, http.StatusBadRequest, "City not found")
		default:
			SendError(c, http.StatusInternalServerError, "Failed to fetch weather data")
		}
		return
	}

	c.JSON(http.StatusOK, data)
}
