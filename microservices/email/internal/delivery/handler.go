package delivery

import (
	"context"
	"encoding/json"
	"net/http"

	"email/internal/domain"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

type sender interface {
	Send(ctx context.Context, req domain.SendEmailRequest) error
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
	logger := loggerPkg.From(ctx)

	var req domain.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warn("Invalid request format", "err", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.To == "" || req.Template == "" {
		logger.Warn("Missing required fields", "missing_to", req.To == "", "missing_template", req.Template == "")
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	if err := h.sender.Send(ctx, req); err != nil {
		logger.Error("Email request failed", "error_chain", err.Error())
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	logger.Debug("Email request completed successfully")
	w.WriteHeader(http.StatusOK)
}
