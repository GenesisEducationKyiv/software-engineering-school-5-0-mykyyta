package httpapi

import (
	"net/http"
	"time"

	loggerCtx "weather/pkg/logger"
)

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r)
		dur := time.Since(start)
		loggerCtx.From(r.Context()).Infow(
			"http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", ww.status,
			"duration_ms", dur.Milliseconds(),
		)
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
