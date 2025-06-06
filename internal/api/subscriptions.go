package api

import (
	"net/http"

	"weatherApi/internal/model"

	"github.com/gin-gonic/gin"
)

// listSubscriptionsHandler returns all subscriptions in the database.
// ⚠️ This endpoint exposes sensitive data and is intended for internal or debugging use only.
// DO NOT enable in production without authentication and proper access control.
func listSubscriptionsHandler(c *gin.Context) {
	var subs []model.Subscription

	if err := DB.Find(&subs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscriptions"})
		return
	}

	c.JSON(http.StatusOK, subs)
}
