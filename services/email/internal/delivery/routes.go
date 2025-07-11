package delivery

import "net/http"

func RegisterRoutes(mux *http.ServeMux, handler *EmailHandler) {
	mux.HandleFunc("/api/email/send", handler.Send)
}
