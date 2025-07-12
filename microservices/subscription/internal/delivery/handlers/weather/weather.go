package weather

import (
	"context"
	"errors"
	"net/http"
	"subscription/internal/delivery/handlers/response"
	"subscription/internal/domain"

	"github.com/gin-gonic/gin"
)

type Report struct {
	Temperature float64 `json:"temperature"`
	Humidity    int     `json:"humidity"`
	Description string  `json:"description"`
}

var ErrCityNotFound = errors.New("city not found")

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
		case errors.Is(err, ErrCityNotFound):
			response.SendError(c, http.StatusBadRequest, "City not found")
		default:
			response.SendError(c, http.StatusInternalServerError, "Failed to fetch weather data")
		}
		return
	}

	dto := Report{
		Temperature: data.Temperature,
		Humidity:    data.Humidity,
		Description: data.Description,
	}
	c.JSON(http.StatusOK, dto)
}
