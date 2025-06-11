package main

import (
	"log"
	"weatherApi/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("❌ Application failed: %v", err)
	}
}
