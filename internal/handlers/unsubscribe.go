package handlers

import (
	"context"
	"errors"
	"net/http"

	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
)

type unsubscribeService interface {
	Unsubscribe(ctx context.Context, token string) error
}

type UnsubscribeHandler struct {
	service unsubscribeService
}

func NewUnsubscribeHandler(service unsubscribeService) *UnsubscribeHandler {
	return &UnsubscribeHandler{service: service}
}

func (h *UnsubscribeHandler) Handle(c *gin.Context) {
	token := c.Param("token")

	if err := h.service.Unsubscribe(c.Request.Context(), token); err != nil {
		switch {
		case errors.Is(err, subscription.ErrInvalidToken):
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid token"})
		case errors.Is(err, subscription.ErrSubscriptionNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "Subscription not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Unsubscribed successfully"})
}
