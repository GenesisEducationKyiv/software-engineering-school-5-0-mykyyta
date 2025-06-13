package api

import (
	"net/http"

	"weatherApi/internal/jwtutil"
	"weatherApi/internal/model"

	"github.com/gin-gonic/gin"
)

// unsubscribeHandler marks a subscription as unsubscribed using a secure token.
// The token is parsed to extract the user's email (acts as a form of lightweight authentication).
// This endpoint does not require login — anyone with the token can unsubscribe.
func unsubscribeHandler(c *gin.Context) {
	token := c.Param("token")

	// Parse the token to extract the associated email
	email, err := jwtutil.Parse(token)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		return
	}

	// Retrieve subscription by email
	var sub model.Subscription
	if err := DB.Where("email = ?", email).First(&sub).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Token not found"})
		return
	}

	// If already unsubscribed, return early
	if sub.IsUnsubscribed {
		c.JSON(http.StatusOK, gin.H{"message": "You are already unsubscribed"})
		return
	}

	// Mark subscription as unsubscribed
	sub.IsUnsubscribed = true
	if err := DB.Save(&sub).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unsubscribe"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}
