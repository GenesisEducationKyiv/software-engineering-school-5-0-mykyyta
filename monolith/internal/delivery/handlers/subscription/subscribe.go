package subscription

import (
	"context"
	"errors"
	"monolith/internal/delivery/handlers/response"
	"monolith/internal/domain"
	"monolith/internal/subscription"
	"net/http"

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
	var req SubscribeRequest
	if err := c.ShouldBind(&req); err != nil {
		response.SendError(c, http.StatusBadRequest, "Invalid input")
		return
	}

	freq := domain.Frequency(req.Frequency)
	if !freq.Valid() {
		response.SendError(c, http.StatusBadRequest, "Invalid frequency value")
		return
	}

	err := h.service.Subscribe(c.Request.Context(), req.Email, req.City, freq)
	if err != nil {
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
