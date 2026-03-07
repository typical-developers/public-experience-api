package config

import (
	"sync"

	"github.com/caarlos0/env/v10"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	// The port that the server will be live on.
	Port string `env:"PORT" envDefault:"8080"`

	// The Opencloud API key necessary for accessing Roblox's Opencloud endpoints.
	OpencloudKey string `env:"OPENCLOUD_KEY,required"`
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
