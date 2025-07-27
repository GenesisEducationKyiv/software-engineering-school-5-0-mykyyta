package main

import (
	"os"

	"gateway/internal/app"
	loggerPkg "gateway/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	logger, err := loggerPkg.New("gateway", env)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Errorw("logger sync failed", "err", err)
		}
	}()

	if err := app.Run(logger); err != nil {
		logger.Fatalf("Application failed: %v", err)
	}
}
