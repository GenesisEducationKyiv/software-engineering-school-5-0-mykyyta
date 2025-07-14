package response

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SendError(c *gin.Context, code int, msg string) {
	log.Printf("error: %s", msg) // або прибрати лог, якщо не потрібно
	c.JSON(code, gin.H{"error": msg})
}

func SendSuccess(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}
