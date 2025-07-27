package main

import (
	"os"

	"gateway/internal/app"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	logger, err := loggerPkg.New(loggerPkg.Config{
		Service: "gateway",
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

	if err := app.Run(logger); err != nil {
		logger.Fatalf("Application failed: %v", err)
	}
}
