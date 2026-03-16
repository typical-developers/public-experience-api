package logger

import (
	"os"

	"go.uber.org/zap"
)

func init() {
	var z *zap.Logger
	var err error

	if v := os.Getenv("ENVIRONMENT"); v == "DEVELOPMENT" {
		z, err = zap.NewDevelopment()
	} else {
		z, err = zap.NewProduction()
	}

	if err != nil {
		panic(err)
	}

	zap.ReplaceGlobals(z)
}
