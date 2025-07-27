package logger

import (
	"go.uber.org/zap"
)

func New(service, env string) (*zap.SugaredLogger, error) {
	var core *zap.Logger
	var err error

	if env == "dev" {
		core, err = zap.NewDevelopment()
	} else {
		core, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	return core.Sugar().With("service", service), nil
}
