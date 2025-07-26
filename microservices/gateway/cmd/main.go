package main

import (
	"gateway/internal/app"
	"gateway/pkg/logger"
	"os"
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
