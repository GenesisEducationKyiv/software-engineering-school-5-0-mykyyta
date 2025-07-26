package response

import (
	"net/http"

	"subscription/pkg/logger"

	"github.com/gin-gonic/gin"
)

func SendError(c *gin.Context, code int, msg string) {
	lg := logger.From(c.Request.Context())
	lg.Errorw("handler error", "msg", msg, "code", code)
	c.JSON(code, gin.H{"error": msg})
}

func SendSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
