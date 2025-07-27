package grpcapi

import (
	"context"
	"time"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	"github.com/google/uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const RequestIDKey = "x-request-id"

func LoggingUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		var reqID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get(RequestIDKey); len(values) > 0 {
				reqID = values[0]
			}
		}
		if reqID == "" {
			reqID = uuid.New().String()
		}

		logger := loggerPkg.From(ctx).With("request_id", reqID)
		ctx = loggerPkg.With(ctx, logger)

		grpc.SetHeader(ctx, metadata.Pairs(RequestIDKey, reqID))

		start := time.Now()
		resp, err = handler(ctx, req)
		duration := time.Since(start)

		st := status.Convert(err)
		code := st.Code()
		codeString := code.String()

		logFields := []interface{}{
			"method", info.FullMethod,
			"status", codeString,
			"duration_ms", duration.Milliseconds(),
		}

		if code == codes.Internal || code == codes.Unavailable || code == codes.DataLoss {
			logger.Errorw("grpc request failed", logFields...)
		} else if code != codes.OK {
			logger.Warnw("grpc request client error", logFields...)
		} else if duration > 1000*time.Millisecond {
			logger.Warnw("slow grpc request", logFields...)
		} else {
			logger.Infow("grpc request", logFields...)
		}

		return resp, err
	}
}
