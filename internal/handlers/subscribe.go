package handlers

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

func (h SubscribeHandler) Handle(c *gin.Context) {
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		SendError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req.Email, req.City, req.Frequency)
	if err != nil {
		switch {
		case errors.Is(err, subscription.ErrCityNotFound):
			SendError(c, http.StatusBadRequest, "City not found")
		case errors.Is(err, subscription.ErrEmailAlreadyExists):
			SendError(c, http.StatusConflict, "Email already subscribed")
		default:
			SendError(c, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	SendSuccess(c, "Subscription successful. Confirmation email sent.")
}
