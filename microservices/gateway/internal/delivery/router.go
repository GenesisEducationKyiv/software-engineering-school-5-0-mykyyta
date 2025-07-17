package delivery

import (
	"log"
	"net/http"

	"gateway/internal/middleware"
)

func SetupRoutes(handler *SubscriptionHandler, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthCheck)
	mux.HandleFunc("/api/subscribe", handler.Subscribe)
	mux.HandleFunc("/api/confirm/", handler.Confirm)
	mux.HandleFunc("/api/unsubscribe/", handler.Unsubscribe)
	mux.HandleFunc("/api/weather", handler.GetWeather)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		apiInfo := []byte(`{
		"service": "api-gateway",
		"version": "1.0.0",
		"endpoints": {
			"health": "GET /health",
			"subscribe": "POST /api/subscribe",
			"confirm": "GET /api/confirm/:token",
			"unsubscribe": "GET /api/unsubscribe/:token",
			"weather": "GET /api/weather?city=<city>"
			}
		}`)

		if _, err := w.Write(apiInfo); err != nil {
			logger.Printf("Failed to write API info response: %v", err)
		}
	})

	var finalHandler http.Handler = mux
	finalHandler = middleware.CORS()(finalHandler)
	finalHandler = middleware.Logging(logger)(finalHandler)

	return finalHandler
}
