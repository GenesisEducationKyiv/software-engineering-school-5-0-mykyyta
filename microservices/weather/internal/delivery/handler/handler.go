package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"weather/internal/domain"
)

type WeatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type WeatherHandler struct {
	service WeatherService
}

func NewWeatherHandler(service WeatherService) *WeatherHandler {
	return &WeatherHandler{service: service}
}

func (h *WeatherHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, `{"error":"city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	report, err := h.service.GetWeather(r.Context(), city)
	if err != nil {
		if errors.Is(err, domain.ErrCityNotFound) {
			http.Error(w, `{"error":"city not found"}`, http.StatusNotFound)
			return
		}

		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"city":        city,
		"temperature": report.Temperature,
		"humidity":    report.Humidity,
		"description": report.Description,
	}

	writeJSON(w, http.StatusOK, resp)
}

func (h *WeatherHandler) ValidateCity(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	if city == "" {
		http.Error(w, `{"error":"city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	valid, err := h.service.CityIsValid(r.Context(), city)
	if err != nil {
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"city":  city,
		"valid": valid,
	}

	writeJSON(w, http.StatusOK, resp)
}

// helper.
func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}
