package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"weather/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type weatherService interface {
	GetWeather(ctx context.Context, city string) (domain.Report, error)
	CityIsValid(ctx context.Context, city string) (bool, error)
}

type Handler struct {
	ws weatherService
}

func NewHandler(ws weatherService) *Handler {
	return &Handler{ws: ws}
}

func (h *Handler) GetWeather(w http.ResponseWriter, r *http.Request) {
	city, err := getQueryParam(r, "city")
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Error("missing city query parameter", "error", err)
		http.Error(w, `{"error":"city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	report, err := h.ws.GetWeather(r.Context(), city)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		if errors.Is(err, domain.ErrCityNotFound) {
			logger.Warn("city not found", "city", city)
			http.Error(w, `{"error":"city not found"}`, http.StatusNotFound)
			return
		}
		logger.Error("failed to get weather", "city", city, "error", err)
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

func (h *Handler) ValidateCity(w http.ResponseWriter, r *http.Request) {
	city, err := getQueryParam(r, "city")
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Error("missing city query parameter", "error", err)
		http.Error(w, `{"error":"city query parameter is required"}`, http.StatusBadRequest)
		return
	}

	valid, err := h.ws.CityIsValid(r.Context(), city)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		if errors.Is(err, domain.ErrCityNotFound) {
			logger.Warn("city not found", "city", city)
			http.Error(w, `{"error":"city not found"}`, http.StatusNotFound)
			return
		}
		logger.Error("failed to validate city", "city", city, "error", err)
		http.Error(w, `{"error":"`+err.Error()+`"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"city":  city,
		"valid": valid,
	}

	writeJSON(w, http.StatusOK, resp)
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func getQueryParam(r *http.Request, param string) (string, error) {
	value := r.URL.Query().Get(param)
	if value == "" {
		return "", fmt.Errorf("missing query parameter: %s", param)
	}
	return value, nil
}
