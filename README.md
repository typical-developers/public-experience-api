# Public Experience API
A publicly accessible API to get data in experiences.

# Prerequesties
- Go (1.26.0)
- A running Redis instance

# Development
Before you can run everything in development mode, you will need to [`air`](https://github.com/air-verse/air), a CLI util for live reloading on code changes.

## Environment
Each service will use different environmental variables. You can refer to the `.env.example` file for examples on variables in each avaiable `cmd` directory.

## API
To run the API in development mode, enter the `/cmd/public/` directory, and run `air`.