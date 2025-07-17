package delivery

import (
	"api-gateway/internal/adapter/subscription"
	"context"
	"encoding/json"
	"net/http"
	"strings"
)

type GatewayService interface {
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
	gatewayService GatewayService
	responseWriter ResponseWriter
	logger         Logger
}

func NewSubscriptionHandler(
	gatewayService GatewayService,
	responseWriter ResponseWriter,
	logger Logger,
) *SubscriptionHandler {
	return &SubscriptionHandler{
		gatewayService: gatewayService,
		responseWriter: responseWriter,
		logger:         logger,
	}
}

// Тільки HTTP: парсинг JSON + виклик service
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req subscription.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode request: %v", err)
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Invalid JSON", "")
		return
	}

	// Делегуємо всю логіку service
	resp, err := h.gatewayService.Subscribe(r.Context(), req)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

// Тільки HTTP: парсинг URL + виклик service
func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/confirm/")
	if token == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Token required", "")
		return
	}

	// Делегуємо всю логіку service
	resp, err := h.gatewayService.Confirm(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

// Тільки HTTP: парсинг URL + виклик service
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	token := strings.TrimPrefix(r.URL.Path, "/api/unsubscribe/")
	if token == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Token required", "")
		return
	}

	// Делегуємо всю логіку service
	resp, err := h.gatewayService.Unsubscribe(r.Context(), token)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

// Тільки HTTP: парсинг query params + виклик service
func (h *SubscriptionHandler) GetWeather(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.responseWriter.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		h.responseWriter.WriteError(w, http.StatusBadRequest, "City required", "")
		return
	}

	// Делегуємо всю логіку service
	resp, err := h.gatewayService.GetWeather(r.Context(), city)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.responseWriter.WriteSuccess(w, resp)
}

// Тільки HTTP: мапінг service errors до HTTP статусів
func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "security validation failed"):
		h.responseWriter.WriteError(w, http.StatusBadRequest, "Security validation failed", "")
	case strings.Contains(errStr, "service failed"):
		h.responseWriter.WriteError(w, http.StatusBadGateway, "Service unavailable", "")
	case strings.Contains(errStr, "context deadline exceeded"):
		h.responseWriter.WriteError(w, http.StatusGatewayTimeout, "Gateway timeout", "")
	default:
		h.responseWriter.WriteError(w, http.StatusInternalServerError, "Internal server error", "")
	}
}
