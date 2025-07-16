// internal/router/router.go
package router

import (
	"log"
	"net/http"

	"api-gateway/internal/client"
	"api-gateway/internal/handler"
	"api-gateway/internal/middleware"
)

func SetupRoutes(subscriptionClient *client.SubscriptionClient, logger *log.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", handler.HealthCheck)

	subscriptionHandler := handler.NewSubscriptionHandler(subscriptionClient, logger)

	mux.HandleFunc("/api/subscription/subscribe", subscriptionHandler.Subscribe)
	mux.HandleFunc("/api/subscription/confirm", subscriptionHandler.Confirm)
	mux.HandleFunc("/api/subscription/unsubscribe", subscriptionHandler.Unsubscribe)

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
            "service": "api-gateway",
            "version": "1.0.0",
            "endpoints": {
                "health": "GET /health",
                "subscribe": "POST /api/subscription/subscribe",
                "confirm": "POST /api/subscription/confirm",
                "unsubscribe": "POST /api/subscription/unsubscribe"
            }
        }`))
	})

	var finalHandler http.Handler = mux
	finalHandler = middleware.CORS()(finalHandler)
	finalHandler = middleware.Logging(logger)(finalHandler)

	return finalHandler
}
