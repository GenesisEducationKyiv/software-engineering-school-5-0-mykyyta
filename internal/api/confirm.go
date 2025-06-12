package api

import (
	"net/http"
	"weatherApi/internal/subscription"

	"weatherApi/internal/jwtutil"
	"weatherApi/internal/scheduler"

	"github.com/gin-gonic/gin"
)

// confirmHandler validates the token and marks the subscription as confirmed.
func confirmHandler(c *gin.Context) {
	token := c.Param("token")

	email, err := jwtutil.Parse(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	var sub subscription.Subscription
	if err := DB.Where("email = ?", email).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found / Subscription not found"})
		return
	}

	if sub.IsConfirmed {
		c.JSON(http.StatusOK, gin.H{"message": "Subscription already confirmed"})
		return
	}

	sub.IsConfirmed = true
	DB.Save(&sub)

	if err := scheduler.ProcessSubscription(c.Request.Context(), sub); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send weather forecast email"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
}
