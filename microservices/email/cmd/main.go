package main

import (
	"os"

	"email/internal/app"
	"email/pkg/logger"
)

func main() {
	env := os.Getenv("ENV")
	lg, err := logger.New("email", env)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := lg.Sync(); err != nil {
			lg.Errorw("logger sync failed", "err", err)
		}
	}()

	lg.Infow("starting service", "env", env)

	if err := app.Run(lg); err != nil {
		lg.Fatalw("service crashed", "err", err)
	}
}
