package main

import (
	"os"

	"email/internal/app"

	loggerPkg "github.com/GenesisEducationKyiv/software-engineering-school-5-0-mykyyta/microservices/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	if env == "" {
		env = "production"
	}
	logger, err := loggerPkg.New(loggerPkg.Config{
		Service: "email",
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

	logger.Info("starting service", "env", env)

	if err := app.Run(logger); err != nil {
		logger.Fatal("service crashed", "err", err)
	}
}
