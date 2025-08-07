package main

import (
	"os"

	"weather/internal/app"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}

	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	logger, err := loggerPkg.New(loggerPkg.Config{
		Service: "weather",
		Env:     env,
		Level:   logLevel,
	})
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := logger.Sync(); err != nil {
			logger.Error("logger sync failed", "err", err)
		}
	}()

	logger.Info("starting service", "env", env)

	if err := app.Run(logger); err != nil {
		logger.Fatal("service crashed", "err", err)
	}
}
