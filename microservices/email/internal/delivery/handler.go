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

	logger.Infow("Received email send request")

	var req domain.SendEmailRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger.Warnw("Invalid JSON in request body", "err", err)
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	if req.To == "" || req.Template == "" {
		logger.Warnw("Missing required fields", "to", req.To, "template", req.Template)
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	logger.Infow("Processing email send request", "to", req.To, "template", req.Template)

	if err := h.sender.Send(ctx, req); err != nil {
		logger.Errorw("Failed to send email", "to", req.To, "template", req.Template, "err", err)
		http.Error(w, "failed to send email", http.StatusInternalServerError)
		return
	}

	logger.Infow("Email send request completed successfully", "to", req.To, "template", req.Template)
	w.WriteHeader(http.StatusOK)
}
