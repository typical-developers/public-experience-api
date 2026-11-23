# Public Experience API
The official, publcily accessible API for data fetching from Typical Developers' Roblox experiences.

## Development
### Prerequisites
- Go 1.26.0
- Redis
- An S3-compatible bucket

### Configuration
All services read a single `.env` file at the repo root. Copy `.env.example` to `.env` and fill it in.
| Variable | Required | Default | Description |
| --- | --- | --- | --- |
| `ENVIRONMENT` | | `production` | `production` or `development`. Development points at the staging experience. |
| `LOG_LEVEL` | | `info` / `debug` | Defaults to `debug` in development, `info` otherwise |
| `OPENCLOUD_KEY` | ✅ | | Roblox Open Cloud API key |
| `ASSET_BASE_URL` | | `http://localhost:8080` | Base URL used to build media URLs in responses (API host or CDN) |
| `LOG_WEBHOOK_URL` | | | Webhook for error and panic logging |
| `OAKLANDS_ALWAYS_EXECUTE_IN_LATEST` | | `false` | _TODO: describe_ |
| `S3_ENDPOINT` | ✅ | | S3-compatible endpoint |
| `S3_REGION` | | | Bucket region |
| `S3_BUCKET` | ✅ | | Bucket name |
| `S3_ACCESS_KEY` | ✅ | | Access key |
| `S3_SECRET_ACCESS_KEY` | ✅ | | Secret key |
| `ASSETS_CONCURRENCY` | | `8` | How many Roblox assets are fetched and uploaded at once |
| `ASSETS_COOLDOWN` | | `1s` | How long to wait before each request to Roblox while fetching assets |
| `ASSETS_MAX_BYTES` | | `52428800` | Largest file, in bytes, downloaded for an asset |
| `REDIS_HOST` | | `localhost` | Redis host |
| `REDIS_PORT` | | `6379` | Redis port |
| `REDIS_PASSWORD` | | | Redis password |
| `REDIS_DB` | | `0` | Redis database index |

### Running with Docker Compose
Docker Compose is used to test and deploy.
```sh
docker compose up --build
```

### Generating `scalar` Docs
The API reference is generated from handler annotations with [swag](https://github.com/swaggo/swag).
```sh
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -g cmd/public/main.go -o docs
```

## Architecture
This project consists of two different services to sync and expose experience data. A Redis instance is hosted to store data received from workers, which is then read and served by the API. Roblox images referenced by that data are hosted in an S3-compatible bucket, so responses can link to them directly. [`scalarapi`](https://scalar.com/) is used to host documentation for the API.
| Service   | Role                                                                                                                       |
| --------- | -------------------------------------------------------------------------------------------------------------------------- |
| `public`  | The HTTP API, exposed on port 8080. Serves hosted images under `/assets/` and fills in their `url` in responses.            |
| `workers` | Jobs that sync experience data, utilizing Roblox's OpenCloud LuauExecution APIs, and host the Roblox images it references. |

```mermaid
flowchart LR
    RB["Roblox<br/>Open Cloud · Thumbnails · CDN"]
    W[workers]
    R[(Redis)]
    S3[(S3)]
    API[public API]
    C[Clients]

    RB -->|data and images| W
    W -->|data and asset manifest| R
    W -->|images| S3
    R --> API
    S3 -->|/assets/| API
    API --> C
```

### Hosting assets
Once Oaklands has been synced, the `SyncRobloxAssets` job runs every minute and only does work for new images and ones that are due for another attempt.
1. Images referenced by newsletters, the minimap and the index are collected. Hidden newsletters are skipped, and videos are not hosted.
2. Images that are already in the bucket are stored without being fetched again.
3. The rest are requested from the public thumbnails API in batches of 20. Newsletter banners use `1440x456`, newsletter content uses `768x432` and everything else uses `700x700`. Thumbnails that are still being generated are requested again at the end of the run.
4. Each image is downloaded, uploaded to `roblox/<asset id>/<hash>.<extension>`, and stored in the `roblox:asset` manifest in Redis.
5. The public API keeps the hosted assets in memory, refreshing when the workers publish an update, and uses them to fill in `url` next to each `asset_id`.

Images that have not been referenced for 7 days are deleted.
