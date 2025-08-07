package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const (
	RequestIDKey    = "requestID"
	RequestIDHeader = "X-Request-ID"
)

func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get(RequestIDHeader)
			if reqID == "" {
				reqID = uuid.NewString()
			}
			ctx := context.WithValue(r.Context(), RequestIDKey, reqID)
			w.Header().Set(RequestIDHeader, reqID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(RequestIDKey); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
