package middleware

import (
	"time"

	loggerPkg "subscription/pkg/logger"

	"github.com/gin-gonic/gin"
)

func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		dur := time.Since(start)
		logger := loggerPkg.From(c.Request.Context())
		logger.Infow(
			"http request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", dur.Milliseconds(),
		)
	}
}
