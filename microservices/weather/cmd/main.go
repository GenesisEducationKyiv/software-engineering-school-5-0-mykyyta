package main

import (
	"weather/internal/app"
	"weather/internal/infra"
)

func main() {
	logg := infra.NewLogger("logs/app.log")

	if err := app.Run(logg); err != nil {
		logg.Fatalf("weather service failed: %v", err)
	}
}
