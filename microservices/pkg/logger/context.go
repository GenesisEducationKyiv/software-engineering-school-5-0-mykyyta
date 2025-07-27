package logger

import (
	"context"

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
