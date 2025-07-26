package httpapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/api/weather", loggingMiddleware(http.HandlerFunc(h.GetWeather)))
	mux.Handle("/api/weather/validate", loggingMiddleware(http.HandlerFunc(h.ValidateCity)))
	mux.Handle("/metrics", promhttp.Handler())
}
