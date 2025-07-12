package weather

import (
	"context"
	"errors"
	"net/http"
	"weatherApi/monolith/internal/delivery/handlers/response"
	"weatherApi/monolith/internal/domain"
	"weatherApi/monolith/internal/weather"

	"github.com/gin-gonic/gin"
)

type WeatherResponse struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

type weatherCurrent interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
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
		response.SendError(c, http.StatusBadRequest, "City is required")
		return
	}

	data, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		switch {
		case errors.Is(err, weather.ErrCityNotFound):
			response.SendError(c, http.StatusBadRequest, "City not found")
		default:
			response.SendError(c, http.StatusInternalServerError, "Failed to fetch weather data")
		}
		return
	}

	dto := WeatherResponse{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}
	c.JSON(http.StatusOK, dto)
}
