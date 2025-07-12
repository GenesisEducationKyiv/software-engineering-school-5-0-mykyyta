package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"subscription/internal/domain"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// --- mock service ---

type mockWeatherService struct {
	getWeatherFunc func(ctx context.Context, city string) (
		domain.Report, error)
}

func (m *mockWeatherService) GetWeather(ctx context.Context, city string) (domain.Report, error) {
	return m.getWeatherFunc(ctx, city)
}

// --- setup router ---

func setupWeatherRouter(service weatherCurrent) *gin.Engine {
	handler := NewWeatherCurrent(service)
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	r.GET("/api/weather", handler.Handle)
	return r
}

// --- test cases ---

func TestWeatherHandler(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := &mockWeatherService{
			getWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{
					Temperature: 21.5,
					Humidity:    60,
					Description: "Sunny",
				}, nil
			},
		}
		router := setupWeatherRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/weather?city=Kyiv", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, `{
			"temperature": 21.5,
			"humidity": 60,
			"description": "Sunny"
		}`, w.Body.String())
	})

	t.Run("MissingCity", func(t *testing.T) {
		service := &mockWeatherService{} // won't be called
		router := setupWeatherRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/weather", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"City is required"}`, w.Body.String())
	})

	t.Run("CityNotFound", func(t *testing.T) {
		service := &mockWeatherService{
			getWeatherFunc: func(ctx context.Context, city string) (domain.Report, error) {
				return domain.Report{}, ErrCityNotFound
			},
		}

		router := setupWeatherRouter(service)

		req := httptest.NewRequest(http.MethodGet, "/api/weather?city=Atlantis", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		assert.JSONEq(t, `{"error":"City not found"}`, w.Body.String())
	})
}
