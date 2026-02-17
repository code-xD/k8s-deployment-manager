package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New returns a new development logger (human-readable, debug level).
// Use NewProduction in production for JSON output and info level.
func New() *zap.Logger {
	return NewDevelopment()
}

// NewDevelopment returns a logger with human-readable console output and debug level.
func NewDevelopment() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	logger, err := config.Build()
	if err != nil {
		panic("failed to build development logger: " + err.Error())
	}
	return logger
}

// NewProduction returns a logger with JSON encoding and info level, suitable for production.
func NewProduction() *zap.Logger {
	logger, err := zap.NewProduction()
	if err != nil {
		panic("failed to build production logger: " + err.Error())
	}
	return logger
}
