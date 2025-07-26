package main

import (
	"gateway/internal/app"
	"gateway/internal/infra"
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
		_ = lg.Sync()
	}()

	if err := app.Run(lg); err != nil {
		lg.Fatalf("Application failed: %v", err)
	}
}
