package logger

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// createLogger will look for ENVIRONMENT.
// If the environment is DEVELOPMENT, it will use NewDevelopment.
// Otherwise, will use NewProduction.
func createLogger() *zap.Logger {
	var z *zap.Logger

	environment := strings.ToLower(os.Getenv("ENVIRONMENT"))

	if environment == "development" {
		z = zap.Must(zap.NewDevelopment())
	} else {
		z = zap.Must(zap.NewProduction())
	}

	return z
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

	config := zap.Config{
		Level: zap.NewAtomicLevelAt(getLogLevel()),
	}

	zap.Must(config.Build())
	zap.ReplaceGlobals(logger)
}
