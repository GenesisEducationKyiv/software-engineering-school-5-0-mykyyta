package handler

import (
	"context"
	"errors"
	"net/http"

	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
)

type subscribeService interface {
	Subscribe(ctx context.Context, email, city, frequency string) error
}

type SubscribeHandler struct {
	service subscribeService
}

func NewSubscribeHandler(service subscribeService) *SubscribeHandler {
	return &SubscribeHandler{service: service}
}

type SubscribeRequest struct {
	Email     string `form:"email" binding:"required,email"`
	City      string `form:"city" binding:"required"`
	Frequency string `form:"frequency" binding:"required,oneof=daily hourly"`
}

func (h *SubscribeHandler) Handle(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req.Email, req.City, req.Frequency)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrCityNotFound):
			c.JSON(http.StatusBadRequest, gin.H{"error": "City not found"})
		case errors.Is(err, subscription.ErrEmailAlreadyExists):
			c.JSON(http.StatusConflict, gin.H{"error": "Email already subscribed"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Something went wrong"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Subscription successful. Confirmation email sent."})
}
