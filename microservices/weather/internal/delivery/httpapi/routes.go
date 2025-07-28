package httpapi

import (
	"net/http"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func RegisterRoutes(mux *http.ServeMux, h *Handler, logger *loggerPkg.Logger) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.Handle("/api/weather", loggingMiddleware(logger)(http.HandlerFunc(h.GetWeather)))
	mux.Handle("/api/weather/validate", loggingMiddleware(logger)(http.HandlerFunc(h.ValidateCity)))
	mux.Handle("/metrics", promhttp.Handler())
}
