module github.com/typical-developers/public-experience-api

go 1.26.0

require (
	github.com/caarlos0/env/v10 v10.0.0
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/cors v1.2.2
	github.com/joho/godotenv v1.5.1
	github.com/redis/go-redis/v9 v9.18.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/typical-developers/goblox v0.0.0-20260312234757-5a027d43801f
	golang.org/x/sync v0.12.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/klauspost/cpuid/v2 v2.2.10 // indirect
	github.com/stretchr/testify v1.9.0 // indirect
	go.uber.org/atomic v1.11.0 // indirect
)

// replace github.com/typical-developers/goblox => ../goblox
