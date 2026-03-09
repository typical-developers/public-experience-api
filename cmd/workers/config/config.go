package config

import (
	"sync"

	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// The Opencloud API key necessary for accessing Roblox's Opencloud endpoints.
	OpencloudKey string `env:"OPENCLOUD_KEY,required"`

	// The config for the Redis instance used for ephemeral storage.
	Redis struct {
		Host     string `env:"HOST" envDefault:"localhost"`
		Port     int    `env:"PORT" envDefault:"6379"`
		Password string `env:"PASSWORD"`
		DB       int    `env:"DB" envDefault:"0"`
	} `envPrefix:"REDIS_"`
}

var (
	once sync.Once
	C    Config
)

func setupConfig() {
	if err := env.Parse(&C); err != nil {
		panic(err)
	}
}

func init() {
	once.Do(setupConfig)
}
