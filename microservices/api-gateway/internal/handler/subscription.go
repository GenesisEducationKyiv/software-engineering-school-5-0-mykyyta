package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"

	"api-gateway/internal/client"
)

type SubscriptionHandler struct {
	client *client.SubscriptionClient
	logger *log.Logger
}

func NewSubscriptionHandler(client *client.SubscriptionClient, logger *log.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		client: client,
		logger: logger,
	}
}

// ErrorResponse для consistent error handling
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// Subscribe створює нову підписку
func (h *SubscriptionHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req client.SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode subscribe request: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", "Request body must be valid JSON")
		return
	}

	// Валідація
	if err := h.validateSubscribeRequest(req); err != nil {
		h.logger.Printf("Subscribe validation failed: %v", err)
		h.writeError(w, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	// Виклик subscription service
	resp, err := h.client.Subscribe(r.Context(), req)
	if err != nil {
		h.logger.Printf("Subscribe failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("Successfully subscribed: %s to %s (%s)", req.Email, req.City, req.Frequency)
	h.writeSuccess(w, resp)
}

// Confirm підтверджує підписку
func (h *SubscriptionHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req client.ConfirmRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode confirm request: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", "Request body must be valid JSON")
		return
	}

	// Валідація
	if req.Token == "" {
		h.writeError(w, http.StatusBadRequest, "Validation failed", "Token is required")
		return
	}

	// Виклик subscription service
	resp, err := h.client.Confirm(r.Context(), req)
	if err != nil {
		h.logger.Printf("Confirm failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("Successfully confirmed subscription with token: %s", req.Token)
	h.writeSuccess(w, resp)
}

// Unsubscribe скасовує підписку
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.writeError(w, http.StatusMethodNotAllowed, "Method not allowed", "")
		return
	}

	var req client.UnsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Printf("Failed to decode unsubscribe request: %v", err)
		h.writeError(w, http.StatusBadRequest, "Invalid JSON", "Request body must be valid JSON")
		return
	}

	// Валідація
	if req.Token == "" {
		h.writeError(w, http.StatusBadRequest, "Validation failed", "Token is required")
		return
	}

	// Виклик subscription service
	resp, err := h.client.Unsubscribe(r.Context(), req)
	if err != nil {
		h.logger.Printf("Unsubscribe failed: %v", err)
		h.handleServiceError(w, err)
		return
	}

	h.logger.Printf("Successfully unsubscribed with token: %s", req.Token)
	h.writeSuccess(w, resp)
}

// validateSubscribeRequest валідує дані для підписки
func (h *SubscriptionHandler) validateSubscribeRequest(req client.SubscribeRequest) error {
	if req.Email == "" {
		return fmt.Errorf("email is required")
	}

	if !h.isValidEmail(req.Email) {
		return fmt.Errorf("invalid email format")
	}

	if req.City == "" {
		return fmt.Errorf("city is required")
	}

	if len(req.City) < 2 {
		return fmt.Errorf("city must be at least 2 characters")
	}

	if req.Frequency == "" {
		return fmt.Errorf("frequency is required")
	}

	validFrequencies := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
	}

	if !validFrequencies[req.Frequency] {
		return fmt.Errorf("frequency must be one of: daily, weekly, monthly")
	}

	return nil
}

// isValidEmail базова валідація email
func (h *SubscriptionHandler) isValidEmail(email string) bool {
	return strings.Contains(email, "@") && strings.Contains(email, ".")
}

// handleServiceError обробляє помилки від subscription service
func (h *SubscriptionHandler) handleServiceError(w http.ResponseWriter, err error) {
	errStr := err.Error()

	switch {
	case strings.Contains(errStr, "status: 400"):
		h.writeError(w, http.StatusBadRequest, "Bad request", "Invalid request data")
	case strings.Contains(errStr, "status: 404"):
		h.writeError(w, http.StatusNotFound, "Not found", "Resource not found")
	case strings.Contains(errStr, "status: 409"):
		h.writeError(w, http.StatusConflict, "Conflict", "Resource already exists")
	case strings.Contains(errStr, "status: 500"):
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "Service temporarily unavailable")
	case strings.Contains(errStr, "context deadline exceeded"):
		h.writeError(w, http.StatusGatewayTimeout, "Gateway timeout", "Service request timeout")
	case strings.Contains(errStr, "connection refused"):
		h.writeError(w, http.StatusServiceUnavailable, "Service unavailable", "Subscription service is unavailable")
	default:
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "An unexpected error occurred")
	}
}

// writeError writes error response
func (h *SubscriptionHandler) writeError(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	json.NewEncoder(w).Encode(response)
}

// writeSuccess writes success response
func (h *SubscriptionHandler) writeSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Printf("Failed to encode success response: %v", err)
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "Failed to encode response")
	}
}
