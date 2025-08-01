package main

import (
	"os"

	"gateway/internal/app"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}
	logger, err := loggerPkg.New(loggerPkg.Config{
		Service: "gateway",
		Env:     env,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("logger sync failed", "err", err)
		}
	}()

	if err := app.Run(logger); err != nil {
		logger.Fatal("Application failed: %v", err)
	}
}
