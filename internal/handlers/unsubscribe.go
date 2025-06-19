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

func (h UnsubscribeHandler) Handle(c *gin.Context) {
	token := c.Param("token")

	if err := h.service.Unsubscribe(c.Request.Context(), token); err != nil {
		switch {
		case errors.Is(err, subscription.ErrInvalidToken):
			SendError(c, http.StatusBadRequest, "Invalid token")
		case errors.Is(err, subscription.ErrSubscriptionNotFound):
			SendError(c, http.StatusNotFound, "Subscription not found")
		default:
			SendError(c, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	SendSuccess(c, "Unsubscribed successfully")
}
