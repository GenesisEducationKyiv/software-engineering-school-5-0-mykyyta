package delivery

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"gateway/internal/adapter/subscription"
	loggerPkg "gateway/pkg/logger"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, req subscription.SubscribeRequest) (*subscription.SubscribeResponse, error)
	Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error)
	Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error)
	GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error)
}

type responseWriter interface {
	WriteError(w http.ResponseWriter, statusCode int, error, message string, r *http.Request)
	WriteSuccess(w http.ResponseWriter, data interface{})
}

type SubscriptionHandler struct {
	subscriptionService SubscriptionService
	responseWriter      responseWriter
}

func NewSubscriptionHandler(
	subscriptionService SubscriptionService,
	responseWriter responseWriter,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		responseWriter:      responseWriter,
	}
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "", r)
		return
	}

	var req subscription.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Errorw("Failed to decode subscribe request", "error", err)
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Invalid JSON", "Request body must be valid JSON", r)
		return
	}

	resp, err := h.subscriptionService.Subscribe(r.Context(), req)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Errorw("Subscribe service failed", "error", err)
		h.handleServiceError(w, err, r)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "", r)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	if token == "" || token == r.URL.Path {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Token is required", r)
		return
	}

	resp, err := h.subscriptionService.Confirm(r.Context(), token)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Errorw("Confirm service failed", "error", err)
		h.handleServiceError(w, err, r)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "", r)
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	if token == "" || token == r.URL.Path {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Token is required", r)
		return
	}

	resp, err := h.subscriptionService.Unsubscribe(r.Context(), token)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Errorw("Unsubscribe service failed", "error", err)
		h.handleServiceError(w, err, r)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "", r)
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "City parameter is required", r)
		return
	}

	resp, err := h.subscriptionService.GetWeather(r.Context(), city)
	if err != nil {
		logger := loggerPkg.From(r.Context())
		logger.Errorw("GetWeather service failed", "error", err)
		h.handleServiceError(w, err, r)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error, r *http.Request) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "validation failed"):
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Invalid input data", r)
	case strings.Contains(errStr, "status: 400"):
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Bad request", "Invalid request data", r)
	case strings.Contains(errStr, "status: 404"):
		h.responseWriter.WriteError(w, http.StatusNotFound, "Not found", "Resource not found", r)
	case strings.Contains(errStr, "status: 409"):
		h.responseWriter.WriteError(w, http.StatusConflict, "Conflict", "Resource already exists", r)
	case strings.Contains(errStr, "status: 500"):
		h.responseWriter.WriteError(w, http.StatusInternalServerError, "Internal server error", "Service temporarily unavailable", r)
	case strings.Contains(errStr, "context deadline exceeded"):
		h.responseWriter.WriteError(w, http.StatusGatewayTimeout, "Gateway timeout", "Service request timeout", r)
	case strings.Contains(errStr, "connection refused"):
		h.responseWriter.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", "Subscription service is unavailable", r)
	default:
		h.responseWriter.WriteError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred", r)
	}
}
