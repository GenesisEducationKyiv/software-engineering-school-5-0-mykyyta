package delivery

import (
	"encoding/json"
	"net/http"

	"email/internal/domain"
	"email/pkg/logger"
)

type sender interface {
	Send(req domain.SendEmailRequest) error
}

type EmailHandler struct {
	sender sender
}

func NewEmailHandler(sender sender) *EmailHandler {
	return &EmailHandler{
		sender: sender,
	}
}

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logg := logger.From(ctx)

	var req domain.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.To == "" || req.Template == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if err := h.sender.Send(req); err != nil {
		logg.Errorw("failed to send email", "err", err)
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
