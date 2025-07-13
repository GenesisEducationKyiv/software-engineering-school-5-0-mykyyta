package subscription

import (
	"context"
	"errors"
	"net/http"
	"subscription/internal/delivery/handlers/response"
	"subscription/internal/service"

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
		case errors.Is(err, service.ErrInvalidToken):
			response.SendError(c, http.StatusBadRequest, "Invalid token")
		case errors.Is(err, service.ErrSubscriptionNotFound):
			response.SendError(c, http.StatusNotFound, "Subscription not found")
		default:
			response.SendError(c, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	response.SendSuccess(c, "Subscription confirmed successfully")
}
