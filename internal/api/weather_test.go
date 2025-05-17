// Package api provides test cases for the weather API endpoints
// ensuring correct handling of weather data requests and responses.
package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"weatherApi/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// mockFetchWithStatus is a function variable that allows tests to inject
// different behaviors for weather fetching operations
var mockFetchWithStatus func(city string) (*model.Weather, int, error)

// init replaces the production fetchWeather function with our test mock
func init() {
	fetchWeather = func(city string) (*model.Weather, int, error) {
		return mockFetchWithStatus(city)
	}
}

// setupTestRouterForWeather creates and configures a Gin router instance
// specifically for testing the weather endpoint
func setupTestRouterForWeather() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/api/weather", getWeatherHandler)
	return router
}

// TestWeatherHandler_Success verifies that the weather endpoint returns
// correct weather data with HTTP 200 status when provided with a valid city
func TestWeatherHandler_Success(t *testing.T) {
	mockFetchWithStatus = func(city string) (*model.Weather, int, error) {
		return &model.Weather{
			Temperature: 21.5,
			Humidity:    60,
			Description: "Sunny",
		}, http.StatusOK, nil
	}

	router := setupTestRouterForWeather()
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=Kyiv", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{
		"temperature": 21.5,
		"humidity": 60,
		"description": "Sunny"
	}`, w.Body.String())
}

// TestWeatherHandler_MissingCity verifies that the weather endpoint returns
// an HTTP 400 error when the city parameter is missing from the request
func TestWeatherHandler_MissingCity(t *testing.T) {
	router := setupTestRouterForWeather()
	req := httptest.NewRequest(http.MethodGet, "/api/weather", nil) // без параметра city
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error":"City is required"}`, w.Body.String())
}

// TestWeatherHandler_CityNotFound verifies that the weather endpoint returns
// an HTTP 404 error when the requested city cannot be found
func TestWeatherHandler_CityNotFound(t *testing.T) {
	mockFetchWithStatus = func(city string) (*model.Weather, int, error) {
		return nil, http.StatusNotFound, errors.New("City not found")
	}

	router := setupTestRouterForWeather()
	req := httptest.NewRequest(http.MethodGet, "/api/weather?city=Nowhere", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error":"City not found"}`, w.Body.String())
}
