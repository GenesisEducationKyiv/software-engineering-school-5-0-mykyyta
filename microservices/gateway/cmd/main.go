package main

import (
	"log"

	"api-gateway/internal/app"
	"api-gateway/internal/config"
)

func main() {
	cfg := config.LoadConfig()

	app, err := app.NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to create app: %v", err)
	}

	if err := app.Start(); err != nil {
		log.Fatalf("Failed to start app: %v", err)
	}
}
