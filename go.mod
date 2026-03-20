module github.com/typical-developers/public-experience-api

go 1.26.0

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/cors v1.2.2
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.18.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/typical-developers/discord-webhooks-go/v2 v2.0.2
	github.com/typical-developers/goblox v0.0.0-20260312234757-5a027d43801f
	github.com/urfave/cli/v3 v3.7.0
	go.uber.org/zap v1.27.1
	golang.org/x/sync v0.13.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	golang.org/x/sys v0.35.0 // indirect
)

// replace github.com/typical-developers/goblox => ../goblox
