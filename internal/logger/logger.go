package logger

import (
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	Environment string
	LogLevel    string
}

var once sync.Once

// Build creates a logger from the provided options.
// If the environment is development, it will use NewDevelopmentConfig.
// Otherwise, it will use NewProductionConfig.
func Build(opts Options) (*zap.Logger, error) {
	var config zap.Config

	environment := strings.ToLower(strings.TrimSpace(opts.Environment))

	if environment == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(parseLogLevel(opts.LogLevel))

	return config.Build()
}

// Init builds and installs the global zap logger once.
func Init(opts Options) *zap.Logger {
	var logger *zap.Logger

	once.Do(func() {
		logger = zap.Must(Build(opts))
		zap.ReplaceGlobals(logger)
	})

	if logger == nil {
		logger = zap.L()
	}

	return logger
}

// parseLogLevel converts a string log level into zap's level type.
// Defaults to Info.
func parseLogLevel(level string) zapcore.Level {
	level = strings.ToLower(strings.TrimSpace(level))

	switch level {
	case "debug":
		return zap.DebugLevel
	case "info":
		return zap.InfoLevel
	case "warn":
		return zap.WarnLevel
	case "error":
		return zap.ErrorLevel
	case "dpanic":
		return zap.DPanicLevel
	case "panic":
		return zap.PanicLevel
	case "fatal":
		return zap.FatalLevel
	default:
		return zap.InfoLevel
	}
}
