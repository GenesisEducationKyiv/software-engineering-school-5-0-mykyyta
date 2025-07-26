package logger

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

func With(ctx context.Context, log *zap.SugaredLogger) context.Context {
	return context.WithValue(ctx, ctxKey{}, log)
}

func From(ctx context.Context) *zap.SugaredLogger {
	if v := ctx.Value(ctxKey{}); v != nil {
		if lgr, ok := v.(*zap.SugaredLogger); ok {
			return lgr
		}
	}
	return zap.NewNop().Sugar()
}
