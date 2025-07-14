package delivery

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *EmailHandler) {
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	mux.HandleFunc("/api/email/send", handler.Send)
}
