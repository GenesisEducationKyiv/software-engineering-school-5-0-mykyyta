package delivery

import (
	"context"
	"net/http"
	"time"

	"email/pkg/logger"

	"github.com/google/uuid"
)

type contextKey string

const (
	requestIDKey    contextKey = "requestID"
	requestIDHeader string     = "X-Request-ID"
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

		log := logger.From(r.Context()).With("request_id", reqID)
		ctx := logger.With(r.Context(), log)

		w.Header().Set(requestIDHeader, reqID)

		start := time.Now()
		ww := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(ww, r.WithContext(ctx))
		dur := time.Since(start)

		if ww.status >= 500 {
			logger.From(ctx).Errorw(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		} else {
			logger.From(ctx).Infow(
				"http request",
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.status,
				"duration_ms", dur.Milliseconds(),
			)
		}
	})
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
