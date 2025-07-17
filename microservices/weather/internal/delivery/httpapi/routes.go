package httpapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("api/weather", h.GetWeather)
	mux.HandleFunc("api/weather/validate", h.ValidateCity)
	mux.Handle("/metrics", promhttp.Handler())
}
