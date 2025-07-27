// Package logger provides a shared logger implementation for all microservices
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config holds logger configuration options
type Config struct {
	Service string
	Env     string
	Level   string
}

// New creates a new sugared logger instance with the specified configuration
func New(cfg Config) (*zap.SugaredLogger, error) {
	var core *zap.Logger
	var err error

	if cfg.Env == "dev" || cfg.Env == "development" {
		core, err = zap.NewDevelopment()
	} else {
		core, err = zap.NewProduction()
	}
	if err != nil {
		return nil, err
	}

	// Set log level if specified
	if cfg.Level != "" {
		level, err := zapcore.ParseLevel(cfg.Level)
		if err == nil {
			core = core.WithOptions(zap.IncreaseLevel(level))
		}
	}

	return core.Sugar().With("service", cfg.Service), nil
}

// NewSimple creates a logger with just service and environment (backward compatibility)
func NewSimple(service, env string) (*zap.SugaredLogger, error) {
	return New(Config{
		Service: service,
		Env:     env,
	})
}
