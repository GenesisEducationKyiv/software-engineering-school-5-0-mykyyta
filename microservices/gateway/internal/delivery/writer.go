package delivery

import (
	"encoding/json"
	"net/http"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
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
	logger := loggerPkg.From(r.Context())
	logger.Errorw("HTTP error response", "status", statusCode, "error", error, "message", message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(response)
}

func (j ResponseWriter) WriteSuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(data)
}
