package handlers

import (
	"context"
	"errors"
	"net/http"

	"weatherApi/internal/subscription"

	"github.com/gin-gonic/gin"
)

type confirm interface {
	Confirm(ctx context.Context, token string) error
}

type Confirm struct {
	service confirm
}

func NewConfirm(service confirm) Confirm {
	return Confirm{
		service: service,
	}
}

func (h Confirm) Handle(c *gin.Context) {
	token := c.Param("token")

	if err := h.service.Confirm(c.Request.Context(), token); err != nil {
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

	SendSuccess(c, "Subscription confirmed successfully")
}
