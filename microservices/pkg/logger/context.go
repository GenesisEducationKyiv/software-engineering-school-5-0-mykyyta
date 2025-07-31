package logger

import (
	"context"
	"crypto/sha256"
	"fmt"

	"go.uber.org/zap"
)

type ctxKey struct{}
type requestIDKey struct{}
type correlationIDKey struct{}

// With adds a logger to the context
func With(ctx context.Context, log *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// From retrieves a logger from the context, returns a no-op logger if none exists
func From(ctx context.Context) *Logger {
	if v := ctx.Value(ctxKey{}); v != nil {
		if lgr, ok := v.(*Logger); ok {
			return lgr
		}
	}

	// Return a no-op logger
	nopCore := zap.NewNop()
	return &Logger{
		sugar: nopCore.Sugar(),
		base:  nopCore,
	}
}

// WithRequestID adds a request ID to the context (technical tracing within service)
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

// GetRequestID retrieves a request ID from the context
func GetRequestID(ctx context.Context) string {
	if v := ctx.Value(requestIDKey{}); v != nil {
		if reqID, ok := v.(string); ok {
			return reqID
		}
	}
	return ""
}

// WithCorrelationID adds a correlation ID to the context (business process tracing)
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey{}, correlationID)
}

// GetCorrelationID retrieves a correlation ID from the context
func GetCorrelationID(ctx context.Context) string {
	if v := ctx.Value(correlationIDKey{}); v != nil {
		if corrID, ok := v.(string); ok {
			return corrID
		}
	}
	return ""
}

// HashEmail creates a short hash of an email for logging (privacy-safe)
func HashEmail(email string) string {
	if email == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(email))
	return fmt.Sprintf("user_%x", hash)[:12]
}
