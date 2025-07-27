package middleware

import (
	"net/http"
	"time"

	loggerPkg "gateway/pkg/logger"

	"go.uber.org/zap"
)

func WithLogger(lg *zap.SugaredLogger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := loggerPkg.With(r.Context(), lg)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			next.ServeHTTP(w, r)

			dur := time.Since(start)
			requestID := GetRequestID(r.Context())
			logger := loggerPkg.From(r.Context())
			logger.Infow(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"duration_ms", dur.Milliseconds(),
				"request_id", requestID,
			)
		})
	}
}
