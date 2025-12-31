package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates a new production logger with JSON encoding
func New(serviceName string) (*zap.Logger, error) {
	config := zap.NewProductionConfig()

	// Use ISO8601 time format for better readability
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	return config.Build()
}

// NewDevelopment creates a development logger with console encoding
func NewDevelopment(serviceName string) (*zap.Logger, error) {
	config := zap.NewDevelopmentConfig()

	// Use ISO8601 time format
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Add service name to all logs
	config.InitialFields = map[string]interface{}{
		"service": serviceName,
	}

	return config.Build()
}

// NewFromEnv creates a logger based on LOG_LEVEL environment variable
// Defaults to production logger if not set
func NewFromEnv(serviceName, logLevel string) (*zap.Logger, error) {
	if logLevel == "debug" || logLevel == "development" {
		return NewDevelopment(serviceName)
	}
	return New(serviceName)
}
