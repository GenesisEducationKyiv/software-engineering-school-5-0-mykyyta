package main

import (
	"api-gateway/internal/app"
	"api-gateway/internal/infra"
)

func main() {
	logg := infra.NewLogger("logs/app.log")

	if err := app.Run(logg); err != nil {
		logg.Fatalf("Application failed: %v", err)
	}
}
