package httpapi

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler, logger *zap.SugaredLogger) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/api/weather", loggingMiddleware(logger)(http.HandlerFunc(h.GetWeather)))
	mux.Handle("/api/weather/validate", loggingMiddleware(logger)(http.HandlerFunc(h.ValidateCity)))
	mux.Handle("/metrics", promhttp.Handler())
}
