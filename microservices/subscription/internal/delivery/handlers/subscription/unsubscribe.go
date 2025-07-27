package subscription

import (
	"errors"
	"net/http"

	"subscription/internal/subscription"

	"golang.org/x/net/context"

	"subscription/internal/delivery/handlers/response"
	loggerPkg "subscription/pkg/logger"

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
	logger := loggerPkg.From(c.Request.Context())
	token := c.Param("token")

	if err := h.service.Unsubscribe(c.Request.Context(), token); err != nil {
		logger.Warnw("unsubscribe failed", "token", token, "err", err)
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

	logger.Infow("unsubscribed successfully", "token", token)
	response.SendSuccess(c, "Unsubscribed successfully")
}
