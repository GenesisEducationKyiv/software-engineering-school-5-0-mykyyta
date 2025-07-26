package main

import (
	"os"

	"gateway/internal/app"
	"gateway/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	lg, err := logger.New("gateway", env)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Errorw("logger sync failed", "err", err)
		}
	}()

	if err := app.Run(lg); err != nil {
		lg.Fatalf("Application failed: %v", err)
	}
}
