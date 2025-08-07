package middleware

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
	metricsPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics"
)

const CorrelationIDKey = "X-Correlation-Id"

func RequestLoggingMiddleware(baseLogger *loggerPkg.Logger, metrics *metricsPkg.Metrics, serviceName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := uuid.New().String()

		corrID := c.GetHeader(CorrelationIDKey)
		if corrID == "" {
			corrID = generateCorrelationID(c)
		}

		c.Writer.Header().Set(CorrelationIDKey, corrID)

		contextLogger := baseLogger.With("request_id", reqID, "correlation_id", corrID)

		ctx := loggerPkg.WithRequestID(c.Request.Context(), reqID)
		ctx = loggerPkg.WithCorrelationID(ctx, corrID)
		ctx = loggerPkg.With(ctx, contextLogger)
		c.Request = c.Request.WithContext(ctx)

		method := c.Request.Method
		path := c.Request.URL.Path
		metrics.IncActiveConnections(serviceName, method, path)

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		metrics.DecActiveConnections(serviceName, method, path)

		status := c.Writer.Status()
		statusStr := strconv.Itoa(status)

		metrics.RecordRequest(serviceName, method, path, statusStr, duration)

		if status >= 500 {
			errorType := getErrorType(status)
			metrics.RecordError(serviceName, method, path, statusStr, errorType)

			loggerPkg.From(ctx).Error("http request failed",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		} else if status >= 400 {
			errorType := getErrorType(status)
			metrics.RecordError(serviceName, method, path, statusStr, errorType)

			loggerPkg.From(ctx).Warn("http request client error",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		} else if duration > 1000*time.Millisecond {
			loggerPkg.From(ctx).Warn("slow http request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		} else {
			loggerPkg.From(ctx).Info("http request",
				"method", method,
				"path", path,
				"status", status,
				"duration_ms", duration.Milliseconds(),
			)
		}
	}
}

func generateCorrelationID(c *gin.Context) string {
	operation := getOperationType(c)
	return operation + "-" + uuid.New().String()[:8]
}

func getOperationType(c *gin.Context) string {
	switch c.Request.URL.Path {
	case "/subscribe":
		return "sub"
	case "/unsubscribe":
		return "unsub"
	case "/confirm":
		return "confirm"
	case "/weather":
		return "weather"
	default:
		return "req"
	}
}

func getErrorType(status int) string {
	switch {
	case status >= 500:
		return "server_error"
	case status >= 400:
		return "client_error"
	default:
		return "unknown"
	}
}
