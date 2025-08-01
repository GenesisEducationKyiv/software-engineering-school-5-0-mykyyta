package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

const CorrelationIDKey = "X-Correlation-Id"

func RequestLoggingMiddleware(baseLogger *loggerPkg.Logger) gin.HandlerFunc {
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

		start := time.Now()
		c.Next()
		duration := time.Since(start)

		status := c.Writer.Status()
		logFields := []interface{}{
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", status,
			"duration_ms", duration.Milliseconds(),
		}

		finalLogger := loggerPkg.From(c.Request.Context())

		if status >= 500 {
			finalLogger.Error("http request failed", logFields...)
		} else if status >= 400 {
			finalLogger.Warn("http request client error", logFields...)
		} else if duration > 1000*time.Millisecond {
			finalLogger.Warn("slow http request", logFields...)
		} else {
			finalLogger.Info("http request", logFields...)
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
