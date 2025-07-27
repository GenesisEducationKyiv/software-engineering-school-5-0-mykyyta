package subscription

import (
	"context"
	"errors"
	"net/http"

	"subscription/internal/subscription"

	"subscription/internal/delivery/handlers/response"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"

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
	logger := loggerPkg.From(c.Request.Context())
	token := c.Param("token")

	if err := h.service.Confirm(c.Request.Context(), token); err != nil {
		logger.Warnw("confirm failed", "token", token, "err", err)
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

	logger.Infow("subscription confirmed", "token", token)
	response.SendSuccess(c, "Subscription confirmed successfully")
}
