package httpapi

import (
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	"github.com/google/uuid"
)

const RequestIDKey = "X-Request-Id"

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqID := r.Header.Get(RequestIDKey)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		w.Header().Set(RequestIDKey, reqID)

		logger := loggerPkg.From(r.Context()).With("request_id", reqID)
		ctx := loggerPkg.With(r.Context(), logger)

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
			logger.Errorw("http request failed", logFields...)
		} else if status >= 400 {
			logger.Warnw("http request client error", logFields...)
		} else if duration > 1000*time.Millisecond {
			logger.Warnw("slow http request", logFields...)
		} else {
			logger.Infow("http request", logFields...)
		}
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}
