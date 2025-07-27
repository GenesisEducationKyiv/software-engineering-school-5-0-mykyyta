package main

import (
	"os"

	"subscription/internal/app"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	logger, err := loggerPkg.New(loggerPkg.Config{
		Service: "subscription",
		Env:     env,
	})
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
