package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

const RequestIDKey = "X-Request-Id"

func RequestLoggingMiddleware(baseLogger *zap.SugaredLogger) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqID := c.GetHeader(RequestIDKey)
		if reqID == "" {
			reqID = uuid.New().String()
		}

		c.Writer.Header().Set(RequestIDKey, reqID)

		contextLogger := baseLogger.With("request_id", reqID)

		ctx := loggerPkg.With(c.Request.Context(), contextLogger)
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

		if status >= 500 {
			contextLogger.Errorw("http request failed", logFields...)
		} else if status >= 400 {
			contextLogger.Warnw("http request client error", logFields...)
		} else if duration > 1000*time.Millisecond {
			contextLogger.Warnw("slow http request", logFields...)
		} else {
			contextLogger.Debugw("http request", logFields...)
		}
	}
}
