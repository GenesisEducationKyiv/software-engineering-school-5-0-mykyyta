package httpapi

import (
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	"github.com/google/uuid"
)

const CorrelationIDKey = "X-Correlation-Id"

func loggingMiddleware(baseLogger *loggerPkg.Logger) func(http.Handler) http.Handler {
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

			start := time.Now()
			ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(ww, r.WithContext(ctx))
			duration := time.Since(start)

			status := ww.status
			logFields := []interface{}{
				"method", r.Method,
				"path", r.URL.Path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			}

			if status >= 500 {
				logger.Error("http request failed", logFields...)
			} else if status >= 400 {
				logger.Warn("http request client error", logFields...)
			} else if duration > 1000*time.Millisecond {
				logger.Warn("slow http request", logFields...)
			} else {
				logger.Info("http request", logFields...)
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
