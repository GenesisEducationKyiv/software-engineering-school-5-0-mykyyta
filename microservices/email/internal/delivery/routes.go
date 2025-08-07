package delivery

import (
	"net/http"

	"github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/metrics"
)

func RegisterRoutes(mux *http.ServeMux, handler *EmailHandler, metrics *metrics.Metrics) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	emailHandler := http.HandlerFunc(handler.Send)
	mux.Handle("/api/email/send", RequestMiddleware(metrics)(emailHandler))
}
