// Package logger provides a shared logger implementation for all microservices
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Service string
	Env     string
	Level   string
}

// Logger wraps zap.SugaredLogger with convenient methods
type Logger struct {
	sugar *zap.SugaredLogger
	base  *zap.Logger
}

// New creates a new Logger instance
func New(cfg Config) (*Logger, error) {
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

	sugar := core.Sugar().With("service", cfg.Service)

	return &Logger{
		sugar: sugar,
		base:  core,
	}, nil
}

// WithFields adds structured fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}

	return &Logger{
		sugar: l.sugar.With(args...),
		base:  l.base,
	}
}

// With adds key-value pairs to the logger (convenience method)
func (l *Logger) With(args ...interface{}) *Logger {
	return &Logger{
		sugar: l.sugar.With(args...),
		base:  l.base,
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.sugar.Debugf(msg, args...)
	} else {
		l.sugar.Debug(msg)
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.sugar.Infof(msg, args...)
	} else {
		l.sugar.Info(msg)
	}
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.sugar.Warnf(msg, args...)
	} else {
		l.sugar.Warn(msg)
	}
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.sugar.Errorf(msg, args...)
	} else {
		l.sugar.Error(msg)
	}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	if len(args) > 0 {
		l.sugar.Fatalf(msg, args...)
	} else {
		l.sugar.Fatal(msg)
	}
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.sugar.Sync()
}

// GetZapLogger returns the underlying zap logger for advanced usage
func (l *Logger) GetZapLogger() *zap.Logger {
	return l.base
}

// GetSugaredLogger returns the underlying sugared logger for advanced usage
func (l *Logger) GetSugaredLogger() *zap.SugaredLogger {
	return l.sugar
}

// NewLegacy creates a sugared logger for backward compatibility
// Deprecated: Use New() instead which returns the enhanced Logger
func NewLegacy(cfg Config) (*zap.SugaredLogger, error) {
	logger, err := New(cfg)
	if err != nil {
		return nil, err
	}
	return logger.sugar, nil
}
