package main

import (
	"os"

	"subscription/internal/app"
	loggerPkg "subscription/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	logger, err := loggerPkg.New("subscription", env)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Errorw("logger sync failed", "err", err)
		}
	}()

	logger.Infow("starting service", "env", env)

	if err := app.Run(logger); err != nil {
		logger.Fatalw("service crashed", "err", err)
	}
}
