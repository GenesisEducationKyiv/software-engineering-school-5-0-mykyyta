package delivery

import (
	"net/http"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/google/uuid"
)

const (
	requestIDHeader string = "X-Request-ID"
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
		reqID := r.Header.Get(requestIDHeader)
		if reqID == "" {
			reqID = uuid.NewString()
		}

		logger := loggerPkg.From(r.Context()).With("request_id", reqID)
		ctx := loggerPkg.With(r.Context(), logger)

		w.Header().Set(requestIDHeader, reqID)

		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r.WithContext(ctx))
		dur := time.Since(start)

		if ww.status >= 500 {
			loggerPkg.From(ctx).Errorw(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		} else {
			loggerPkg.From(ctx).Infow(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		}
	})
}
