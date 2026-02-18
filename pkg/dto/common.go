package dto

import "go.uber.org/zap"

var (
	// APICfg holds the global API configuration
	APICfg *APIConfig

	// Log holds the global Zap logger instance
	Log *zap.Logger
)
