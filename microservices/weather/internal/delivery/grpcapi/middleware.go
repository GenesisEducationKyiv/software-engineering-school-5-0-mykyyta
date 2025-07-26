package grpcapi

import (
	"context"
	"time"

	loggerCtx "weather/pkg/logger"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingUnaryServerInterceptor logs gRPC requests with method, status, and duration.
func LoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()
		resp, err = handler(ctx, req)
		dur := time.Since(start)
		st := status.Convert(err)
		code := "OK"
		if st != nil && st.Code().String() != "OK" {
			code = st.Code().String()
		}
		loggerCtx.From(ctx).Infow(
			"grpc request",
			"method", info.FullMethod,
			"status", code,
			"duration_ms", dur.Milliseconds(),
		)
		return resp, err
	}
}
