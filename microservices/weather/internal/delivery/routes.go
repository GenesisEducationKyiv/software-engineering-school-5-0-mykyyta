package delivery

import (
	"net/http"

	"weather/internal/delivery/handler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.WeatherHandler) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("api/weather", h.GetWeather)
	mux.HandleFunc("api/weather/validate", h.ValidateCity)
	mux.Handle("/metrics", promhttp.Handler())
}
