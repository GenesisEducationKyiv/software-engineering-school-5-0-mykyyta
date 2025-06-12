package api

import (
	"context"
	"errors"
	"net/http"
	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
)

type confirmService interface {
	Confirm(ctx context.Context, token string) error
}

type ConfirmHandler struct {
	service confirmService
}

func NewConfirmHandler(service confirmService) *ConfirmHandler {
	return &ConfirmHandler{
		service: service,
	}
}

func (h *ConfirmHandler) Handle(c *gin.Context) {
	token := c.Param("token")

	if err := h.service.Confirm(c.Request.Context(), token); err != nil {
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

	c.JSON(http.StatusOK, gin.H{"message": "Subscription confirmed successfully"})
}
