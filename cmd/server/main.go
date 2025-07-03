package main

import (
	"weatherApi/internal/app"
	"weatherApi/internal/infra"
)

func main() {
	logg := infra.NewLogger("logs/app.log")

	if err := app.Run(logg); err != nil {
		logg.Fatalf("Application failed: %v", err)
	}
}
