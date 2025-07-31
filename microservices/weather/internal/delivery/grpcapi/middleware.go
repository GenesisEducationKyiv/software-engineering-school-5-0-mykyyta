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

const CorrelationIDKey = "x-correlation-id"

func LoggingUnaryServerInterceptor(baseLogger *loggerPkg.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		reqID := uuid.New().String()

		var corrID string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			if values := md.Get(CorrelationIDKey); len(values) > 0 {
				corrID = values[0]
			}
		}
		if corrID == "" {
			corrID = "weather-" + uuid.New().String()[:8]
		}

		logger := baseLogger.With("request_id", reqID, "correlation_id", corrID)

		ctx = loggerPkg.WithRequestID(ctx, reqID)
		ctx = loggerPkg.WithCorrelationID(ctx, corrID)
		ctx = loggerPkg.With(ctx, logger)

		if err := grpc.SetHeader(ctx, metadata.Pairs(
			CorrelationIDKey, corrID,
		)); err != nil {
			logger.Warn("failed to set correlation header", "error", err)
		}

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
			logger.Error("grpc request failed", logFields...)
		} else if code != codes.OK {
			logger.Warn("grpc request client error", logFields...)
		} else if duration > 1000*time.Millisecond {
			logger.Warn("slow grpc request", logFields...)
		} else {
			logger.Info("grpc request", logFields...)
		}

		return resp, err
	}
}
