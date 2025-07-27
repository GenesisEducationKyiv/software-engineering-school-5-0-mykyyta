package middleware

import (
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"go.uber.org/zap"
)

func WithLogger(logger *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := loggerPkg.With(r.Context(), logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status  int
	written bool
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.written {
		rw.status = code
		rw.written = true
	}
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(data []byte) (int, error) {
	if !rw.written {
		rw.status = http.StatusOK
		rw.written = true
	}
	return rw.ResponseWriter.Write(data)
}

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := GetRequestID(r.Context())

			ww := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
				written:        false,
			}

			next.ServeHTTP(ww, r)

			dur := time.Since(start)
			logger := loggerPkg.From(r.Context())

			if ww.status >= 500 {
				logger.Errorw("HTTP request failed",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.status,
					"duration_ms", dur.Milliseconds(),
					"request_id", requestID)
			} else if ww.status >= 400 {
				logger.Warnw("HTTP client error",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.status,
					"duration_ms", dur.Milliseconds(),
					"request_id", requestID)
			} else if dur > 1000*time.Millisecond {
				logger.Warnw("Slow HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.status,
					"duration_ms", dur.Milliseconds(),
					"request_id", requestID)
			} else {
				logger.Debugw("HTTP request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", ww.status,
					"duration_ms", dur.Milliseconds(),
					"request_id", requestID)
			}
		})
	}
}
