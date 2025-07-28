package config

import "os"

var (
	OpencloudAPIKey string

	RedisHost     string
	RedisPassword string
	RedisDB       string

	PostgresURL string
)

func Load() {
	var ok bool

	if OpencloudAPIKey, ok = os.LookupEnv("OPENCLOUD_API_KEY"); !ok {
		panic("OPENCLOUD_API_KEY is not set")
	}

	if RedisHost, ok = os.LookupEnv("REDIS_HOST"); !ok {
		panic("REDIS_HOST is not set")
	}

	if RedisPassword, ok = os.LookupEnv("REDIS_PASSWORD"); !ok {
		panic("REDIS_PASSWORD is not set")
	}

	if RedisDB, ok = os.LookupEnv("REDIS_DB"); !ok {
		panic("REDIS_DB is not set")
	}

	if PostgresURL, ok = os.LookupEnv("POSTGRES_URL"); !ok {
		panic("POSTGRES_URL is not set")
	}
}
