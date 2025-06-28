package main

import (
	"log"
	"os"

	"weatherApi/internal/weather/metrics"

	"weatherApi/internal/app"
)

func main() {
	metrics.Register()

	logFile, err := os.OpenFile("logs/app.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		log.Fatalf("cannot open log file: %v", err)
	}
	defer func() {
		if err := logFile.Close(); err != nil {
			log.Fatalf("failed to close log file: %v", err)
		}
	}()

	logger := log.New(logFile, "", log.LstdFlags)

	if err := app.Run(logger); err != nil {
		logger.Fatalf("Application failed: %v", err)
	}
}
