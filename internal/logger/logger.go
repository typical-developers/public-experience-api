package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// createLogger will look for ENVIRONMENT.
// If the environment is development, it will use NewDevelopmentConfig.
// Otherwise, it will use NewProductionConfig.
func createLogger() *zap.Logger {
	var config zap.Config

	environment := strings.ToLower(os.Getenv("ENVIRONMENT"))

	if environment == "development" {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	config.Level = zap.NewAtomicLevelAt(getLogLevel())

	return zap.Must(config.Build())
}

// getLogLevel will look for LOG_LEVEL in the environment.
// defaults to Info.
func getLogLevel() zapcore.Level {
	level := strings.ToLower(os.Getenv("LOG_LEVEL"))

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

func init() {
	logger := createLogger()
	zap.ReplaceGlobals(logger)
}
