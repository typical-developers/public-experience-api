package config

import (
	"strings"

	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// The environment that the server is currently running in.
	// Possible values are "production" and "development"
	Environment string `env:"ENVIRONMENT" envDefault:"production"`
	// The log level that the logger should run at. If no value is provided,
	// it will adjust between `info` and `debug` depending on the environment set.
	LogLevel string `env:"LOG_LEVEL"`

	// The Opencloud API key necessary for accessing Roblox's Opencloud endpoints.
	OpencloudKey string `env:"OPENCLOUD_KEY,required"`

	// The webhook used for logging errors and panics.
	LogWebhookURL *string `env:"LOG_WEBHOOK_URL"`

	Oaklands struct {
		AlwaysExecuteInLatest bool `env:"ALWAYS_EXECUTE_IN_LATEST" envDefault:"false"`
	} `envPrefix:"OAKLANDS_"`

	// The config for the Redis instance used for ephemeral storage.
	Redis struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     int    `env:"PORT" envDefault:"6379"`
		Password string `env:"PASSWORD"`
		DB       int    `env:"DB" envDefault:"0"`
	} `envPrefix:"REDIS_"`
}

var (
	C Config
)

func init() {
	if err := env.Parse(&C); err != nil {
		panic(err)
	}

	if strings.TrimSpace(C.LogLevel) == "" {
		if strings.EqualFold(C.Environment, "development") {
			C.LogLevel = "debug"
		} else {
			C.LogLevel = "info"
		}
	}
}
