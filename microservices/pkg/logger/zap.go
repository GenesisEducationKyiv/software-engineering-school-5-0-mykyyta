// Package logger provides a shared logger implementation for all microservices
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"os"
	"time"
)

type Config struct {
	Service string
	Env     string
	Level   string
}

type Logger struct {
	sugar *zap.SugaredLogger
	base  *zap.Logger
}

func New(cfg Config) (*Logger, error) {
	var logger *zap.Logger
	var err error

	if cfg.Env == "dev" || cfg.Env == "development" {
		logger, err = zap.NewDevelopment()
	} else {
		config := zap.NewProductionConfig()
		config.OutputPaths = []string{"stdout"}
		config.ErrorOutputPaths = []string{"stderr"}

		if cfg.Level != "" {
			if level, parseErr := zapcore.ParseLevel(cfg.Level); parseErr == nil {
				config.Level = zap.NewAtomicLevelAt(level)
			}
		}

		encoder := zapcore.NewJSONEncoder(config.EncoderConfig)
		writeSyncer := zapcore.AddSync(os.Stdout)

		// Core for all levels except Info (no sampling)
		nonInfoCore := zapcore.NewCore(
			encoder,
			writeSyncer,
			zap.LevelEnablerFunc(func(level zapcore.Level) bool {
				return level != zapcore.InfoLevel && level >= config.Level.Level()
			}),
		)

		// Core for Info level only (with sampling)
		infoCore := zapcore.NewSamplerWithOptions(
			zapcore.NewCore(
				encoder,
				writeSyncer,
				zapcore.InfoLevel,
			),
			time.Second, // Sample window
			50,          // First 50 per second
			50,          // Then every 50th
		)

		core := zapcore.NewTee(nonInfoCore, infoCore)
		logger = zap.New(core)
	}

	if err != nil {
		return nil, err
	}

	sugar := logger.Sugar().With("service", cfg.Service)

	return &Logger{
		sugar: sugar,
		base:  logger,
	}, nil
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
	l.sugar.Debugw(msg, args...)
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.sugar.Infow(msg, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, args ...interface{}) {
	l.sugar.Warnw(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.sugar.Errorw(msg, args...)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string, args ...interface{}) {
	l.sugar.Fatalw(msg, args...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.sugar.Sync()
}
