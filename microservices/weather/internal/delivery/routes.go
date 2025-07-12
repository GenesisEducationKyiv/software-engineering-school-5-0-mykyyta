package delivery

import (
	"net/http"

	"weather/internal/delivery/handler"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *handler.WeatherHandler) {
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/weather", h.GetWeather)
	mux.HandleFunc("/weather/validate", h.ValidateCity)
	mux.Handle("/metrics", promhttp.Handler())
}
