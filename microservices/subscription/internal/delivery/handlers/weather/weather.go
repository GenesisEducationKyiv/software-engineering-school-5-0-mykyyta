package weather

import (
	"context"
	"errors"
	"net/http"

	"subscription/internal/delivery/handlers/response"
	"subscription/internal/domain"
	loggerPkg "subscription/pkg/logger"

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
	logger := loggerPkg.From(c.Request.Context())
	city := c.Query("city")
	if city == "" {
		logger.Warnw("city is required for weather", "query", c.Request.URL.RawQuery)
		response.SendError(c, http.StatusBadRequest, "City is required")
		return
	}

	data, err := h.service.GetWeather(c.Request.Context(), city)
	if err != nil {
		logger.Warnw("failed to fetch weather data", "city", city, "err", err)
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

	logger.Infow("weather data fetched", "city", city, "data", dto)
	c.JSON(http.StatusOK, dto)
}
