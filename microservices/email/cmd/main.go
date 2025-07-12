package main

import (
	"email/internal/app"
	"email/internal/infra"
)

func main() {
	logg := infra.NewLogger("logs/app.log")

	if err := app.Run(logg); err != nil {
		logg.Fatalf("email service failed: %v", err)
	}
}
