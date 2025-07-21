package subscription

import (
	"context"
	"errors"
	"net/http"

	"subscription/internal/delivery/handlers/response"
	"subscription/internal/subscription"

	"github.com/gin-gonic/gin"
)

type unsubscribe interface {
	Unsubscribe(ctx context.Context, token string) error
}

type Unsubscribe struct {
	service unsubscribe
}

func NewUnsubscribe(service unsubscribe) Unsubscribe {
	return Unsubscribe{service: service}
}

func (h Unsubscribe) Handle(c *gin.Context) {
	token := c.Param("token")

	if err := h.service.Unsubscribe(c.Request.Context(), token); err != nil {
		switch {
		case errors.Is(err, subscription.ErrInvalidToken):
			response.SendError(c, http.StatusBadRequest, "Invalid token")
		case errors.Is(err, subscription.ErrSubscriptionNotFound):
			response.SendError(c, http.StatusNotFound, "Subscription not found")
		default:
			response.SendError(c, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	response.SendSuccess(c, "Unsubscribed successfully")
}
