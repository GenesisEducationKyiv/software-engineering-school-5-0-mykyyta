package delivery

import (
	"encoding/json"
	"net/http"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

type ResponseWriter struct{}

func NewResponseWriter() ResponseWriter {
	return ResponseWriter{}
}

func (j ResponseWriter) WriteError(w http.ResponseWriter, statusCode int, error, message string, r *http.Request) {
	response := ErrorResponse{
		Error:   error,
		Message: message,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func (j ResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
