package httpapi

import (
	"fmt"
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	metricsPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics"
	"github.com/google/uuid"
)

const CorrelationIDKey = "X-Correlation-Id"

func loggingMiddleware(baseLogger *loggerPkg.Logger, metrics *metricsPkg.Metrics) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := uuid.New().String()

			corrID := r.Header.Get(CorrelationIDKey)
			if corrID == "" {
				corrID = "weather-" + uuid.New().String()[:8]
			}

			w.Header().Set(CorrelationIDKey, corrID)

			logger := baseLogger.With("request_id", reqID, "correlation_id", corrID)

			ctx := loggerPkg.WithRequestID(r.Context(), reqID)
			ctx = loggerPkg.WithCorrelationID(ctx, corrID)
			ctx = loggerPkg.With(ctx, logger)

			method := r.Method
			path := r.URL.Path
			metrics.IncActiveConnections("weather-service", method, path)

			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r.WithContext(ctx))
			duration := time.Since(start)

			metrics.DecActiveConnections("weather-service", method, path)

			status := ww.status
			statusStr := fmt.Sprintf("%d", status)

			metrics.RecordRequest("weather-service", method, path, statusStr, duration)

			if status >= 500 {
				metrics.RecordError("weather-service", method, path, statusStr, "server_error")

				loggerPkg.From(ctx).Error("http request failed",
					"method", method,
					"path", path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)
			} else if status >= 400 {
				metrics.RecordError("weather-service", method, path, statusStr, "client_error")

				loggerPkg.From(ctx).Warn("http request client error",
					"method", method,
					"path", path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)
			} else if duration > 1000*time.Millisecond {
				loggerPkg.From(ctx).Warn("slow http request",
					"method", method,
					"path", path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)
			} else {
				loggerPkg.From(ctx).Info("http request",
					"method", method,
					"path", path,
					"status", status,
					"duration_ms", duration.Milliseconds(),
				)
			}
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
