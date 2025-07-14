package delivery

import (
	"encoding/json"
	"log"
	"net/http"

	"email/internal/domain"
	"email/internal/service"
)

type EmailHandler struct {
	sender service.Sender
	logger *log.Logger
}

func NewEmailHandler(sender service.Sender, logger *log.Logger) *EmailHandler {
	return &EmailHandler{
		sender: sender,
		logger: logger,
	}
}

func (h *EmailHandler) Send(w http.ResponseWriter, r *http.Request) {
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
		h.logger.Printf("failed to send email: %v", err)
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
