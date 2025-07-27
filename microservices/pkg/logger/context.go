package logger

import (
	"context"
	"crypto/sha256"
	"fmt"

	"go.uber.org/zap"
)

type ctxKey struct{}

// With adds a logger to the context
func With(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

// From retrieves a logger from the context, returns a no-op logger if none exists
func From(ctx context.Context) *zap.SugaredLogger {
	if v := ctx.Value(ctxKey{}); v != nil {
		if lgr, ok := v.(*zap.SugaredLogger); ok {
			return lgr
		}
	}
	return zap.NewNop().Sugar()
}

// HashEmail creates a short hash of an email for logging (privacy-safe)
func HashEmail(email string) string {
	if email == "" {
		return ""
	}
	hash := sha256.Sum256([]byte(email))
	return fmt.Sprintf("user_%x", hash)[:12]
}
