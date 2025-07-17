package delivery

import (
	"encoding/json"
	"log"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ResponseWriter struct {
	logger *log.Logger
}

func NewResponseWriter(logger *log.Logger) ResponseWriter {
	return ResponseWriter{
		logger: logger,
	}
}

func (j ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, error, message string) {
	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		j.logger.Printf("Failed to encode error response: %v", err)
	}
}

func (j ResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		j.logger.Printf("Failed to encode success response: %v", err)
		j.WriteError(w, http.StatusInternalServerError, "Internal server error", "Failed to encode response")
	}
}
