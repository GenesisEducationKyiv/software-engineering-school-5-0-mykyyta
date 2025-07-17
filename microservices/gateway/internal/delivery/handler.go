package delivery

import (
	"api-gateway/internal/adapter/subscription"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type SubscriptionService interface {
	Subscribe(ctx context.Context, req subscription.SubscribeRequest) (*subscription.SubscribeResponse, error)
	Confirm(ctx context.Context, token string) (*subscription.ConfirmResponse, error)
	Unsubscribe(ctx context.Context, token string) (*subscription.UnsubscribeResponse, error)
	GetWeather(ctx context.Context, city string) (*subscription.WeatherResponse, error)
}

type ResponseWriter interface {
	WriteError(w http.ResponseWriter, statusCode int, error, message string)
	WriteSuccess(w http.ResponseWriter, data interface{})
}

type Logger interface {
	Printf(format string, v ...interface{})
}

type SubscriptionHandler struct {
	subscriptionService SubscriptionService
	responseWriter      ResponseWriter
	logger              Logger
}

func NewSubscriptionHandler(
	subscriptionService SubscriptionService,
	responseWriter ResponseWriter,
	logger Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		subscriptionService: subscriptionService,
		responseWriter:      responseWriter,
		logger:              logger,
	}
}

func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req subscription.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode subscribe request: %v", err)
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Invalid JSON", "Request body must be valid JSON")
		return
	}

	resp, err := h.subscriptionService.Subscribe(r.Context(), req)
	if err != nil {
		h.logger.Printf("Subscribe service failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	if token == "" || token == r.URL.Path {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Token is required")
		return
	}

	resp, err := h.subscriptionService.Confirm(r.Context(), token)
	if err != nil {
		h.logger.Printf("Confirm service failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	if token == "" || token == r.URL.Path {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Token is required")
		return
	}

	resp, err := h.subscriptionService.Unsubscribe(r.Context(), token)
	if err != nil {
		h.logger.Printf("Unsubscribe service failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "City parameter is required")
		return
	}

	resp, err := h.subscriptionService.GetWeather(r.Context(), city)
	if err != nil {
		h.logger.Printf("GetWeather service failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "validation failed"):
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Validation failed", "Invalid input data")
	case strings.Contains(errStr, "status: 400"):
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Bad request", "Invalid request data")
	case strings.Contains(errStr, "status: 404"):
		h.responseWriter.WriteError(w, http.StatusNotFound, "Not found", "Resource not found")
	case strings.Contains(errStr, "status: 409"):
		h.responseWriter.WriteError(w, http.StatusConflict, "Conflict", "Resource already exists")
	case strings.Contains(errStr, "status: 500"):
		h.responseWriter.WriteError(w, http.StatusInternalServerError, "Internal server error", "Service temporarily unavailable")
	case strings.Contains(errStr, "context deadline exceeded"):
		h.responseWriter.WriteError(w, http.StatusGatewayTimeout, "Gateway timeout", "Service request timeout")
	case strings.Contains(errStr, "connection refused"):
		h.responseWriter.WriteError(w, http.StatusServiceUnavailable, "Service unavailable", "Subscription service is unavailable")
	default:
		h.responseWriter.WriteError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
	}
}
