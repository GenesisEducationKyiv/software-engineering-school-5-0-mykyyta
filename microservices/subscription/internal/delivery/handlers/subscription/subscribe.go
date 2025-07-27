package subscription

import (
	"context"
	"errors"
	"net/http"

	"subscription/internal/domain"

	"subscription/internal/delivery/handlers/response"
	"subscription/internal/subscription"
	loggerPkg "subscription/pkg/logger"

	"github.com/gin-gonic/gin"
)

type subscribe interface {
	Subscribe(ctx context.Context, email, city string, frequency domain.Frequency) error
}

type Subscribe struct {
	service subscribe
}

func NewSubscribe(service subscribe) Subscribe {
	return Subscribe{service: service}
}

type SubscribeRequest struct {
	Email     string `form:"email" binding:"required,email"`
	City      string `form:"city" binding:"required"`
	Frequency string `form:"frequency" binding:"required,oneof=daily hourly"`
}

func (h Subscribe) Handle(c *gin.Context) {
	logger := loggerPkg.From(c.Request.Context())
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		logger.Warnw("invalid subscribe input", "err", err)
		response.SendError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	freq := domain.Frequency(req.Frequency)
	if !freq.Valid() {
		logger.Warnw("invalid frequency value", "value", req.Frequency)
		response.SendError(c, http.StatusBadRequest, "Invalid frequency value")
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req.Email, req.City, freq)
	if err != nil {
		logger.Warnw("subscribe failed", "email", req.Email, "city", req.City, "err", err)
		switch {
		case errors.Is(err, subscription.ErrCityNotFound):
			response.SendError(c, http.StatusBadRequest, "City not found")
		case errors.Is(err, subscription.ErrEmailAlreadyExists):
			response.SendError(c, http.StatusConflict, "Email already subscribed")
		default:
			response.SendError(c, http.StatusInternalServerError, "Something went wrong")
		}
		return
	}

	response.SendSuccess(c, "Subscription successful. Confirmation email sent.")
}
