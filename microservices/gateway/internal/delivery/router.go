package delivery

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"gateway/internal/config"
	"gateway/internal/middleware"
)

func SetupRoutes(handler *SubscriptionHandler, logger *log.Logger, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", HealthCheck)

	handlers := map[string]http.HandlerFunc{
		"Subscribe":   handler.Subscribe,
		"Confirm":     handler.Confirm,
		"Unsubscribe": handler.Unsubscribe,
		"GetWeather":  handler.GetWeather,
	}

	for _, rt := range cfg.Routes {
		hf, ok := handlers[rt.Handler]
		if !ok {
			logger.Printf("unknown handler: %s – skipping", rt.Handler)
			continue
		}
		mux.HandleFunc(rt.Path, hf)
	}

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		if err := json.NewEncoder(w).Encode(generateAPIDocs(cfg)); err != nil {
			logger.Printf("Failed to write API docs: %v", err)
		}
	})

	// middleware
	var finalHandler http.Handler = mux
	finalHandler = middleware.CORS()(finalHandler)
	finalHandler = middleware.Logging(logger)(finalHandler)

	return finalHandler
}

func generateAPIDocs(cfg *config.Config) any {
	type ep struct {
		Method string `json:"method"`
		Path   string `json:"path"`
	}

	endpoints := make(map[string]ep, len(cfg.Routes))
	for _, rt := range cfg.Routes {
		key := strings.TrimPrefix(rt.Path, "/api/")
		endpoints[key] = ep{
			Method: rt.Method,
			Path:   rt.Path,
		}
	}

	return map[string]any{
		"service":   cfg.Service,
		"version":   cfg.Version,
		"endpoints": endpoints,
	}
}
