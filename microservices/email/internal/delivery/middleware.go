package delivery

import (
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	"github.com/google/uuid"
)

const (
	CorrelationIDKey = "X-Correlation-Id"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := uuid.New().String()

		corrID := r.Header.Get(CorrelationIDKey)
		if corrID == "" {
			corrID = "email-" + uuid.New().String()[:8]
		}

		w.Header().Set(CorrelationIDKey, corrID)

		logger := loggerPkg.From(r.Context()).With("request_id", reqID, "correlation_id", corrID)

		ctx := loggerPkg.WithRequestID(r.Context(), reqID)
		ctx = loggerPkg.WithCorrelationID(ctx, corrID)
		ctx = loggerPkg.With(ctx, logger)

		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r.WithContext(ctx))
		dur := time.Since(start)

		if ww.status >= 500 {
			loggerPkg.From(ctx).Error(
				"http request failed",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		} else if ww.status >= 400 {
			loggerPkg.From(ctx).Warn(
				"http request client error",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		} else if dur > 1000*time.Millisecond { // Log slow requests
			loggerPkg.From(ctx).Warn(
				"slow http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		} else {
			loggerPkg.From(ctx).Debug(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		}
	})
}
