package response

import (
	"net/http"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

	"github.com/gin-gonic/gin"
)

func SendError(c *gin.Context, code int, msg string) {
	logger := loggerPkg.From(c.Request.Context())
	logger.Errorw("handler error", "msg", msg, "code", code)
	c.JSON(code, gin.H{"error": msg})
}

func SendSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
