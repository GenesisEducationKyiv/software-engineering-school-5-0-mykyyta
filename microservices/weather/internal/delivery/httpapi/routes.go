package httpapi

import (
	"net/http"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	metricsPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler, logger *loggerPkg.Logger, metrics *metricsPkg.Metrics) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/api/weather", loggingMiddleware(logger, metrics)(http.HandlerFunc(h.GetWeather)))
	mux.Handle("/api/weather/validate", loggingMiddleware(logger, metrics)(http.HandlerFunc(h.ValidateCity)))
	mux.Handle("/metrics", promhttp.Handler())
}
