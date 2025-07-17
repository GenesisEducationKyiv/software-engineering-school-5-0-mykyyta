package response

import (
	"encoding/json"
	"log"
	"net/http"
)

type HTTPResponseWriter interface {
	WriteError(w http.ResponseWriter, statusCode int, error, message string)
	WriteSuccess(w http.ResponseWriter, data interface{})
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type jsonResponseWriter struct {
	logger *log.Logger
}

func NewJSONResponseWriter(logger *log.Logger) HTTPResponseWriter {
	return &jsonResponseWriter{
		logger: logger,
	}
}

func (j *jsonResponseWriter) WriteError(w http.ResponseWriter, statusCode int, error, message string) {
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

func (j *jsonResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		j.logger.Printf("Failed to encode success response: %v", err)
		j.WriteError(w, http.StatusInternalServerError, "Internal server error", "Failed to encode response")
	}
}
