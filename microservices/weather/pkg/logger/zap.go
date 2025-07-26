package logger

import (
	"go.uber.org/zap"
)

func New(service, env string) (*zap.SugaredLogger, error) {
	var core *zap.Logger
	var err error

	if env == "dev" {
		core, err = zap.NewDevelopment() // гарні кольори та stacktrace
	} else {
		core, err = zap.NewProduction() // мінімальний JSON
	}
	if err != nil {
		return nil, err
	}

	return core.Sugar().With("service", service), nil
}
