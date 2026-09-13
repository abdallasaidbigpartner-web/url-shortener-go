# URL Shortener (Go)

![CI](https://github.com/abdallasaidbigpartner-web/url-shortener-go/actions/workflows/ci.yml/badge.svg) ![License](https://img.shields.io/badge/license-MIT-blue.svg)

A URL shortener service demonstrating Go's core strength: concurrent processing via goroutines and channels. Click analytics are recorded asynchronously in a background worker, decoupled from the user-facing redirect - a standard real-world pattern for separating a fast-path (the redirect, which must be instant) from a slow-path (analytics writes, which can happen whenever).

## Architecture

    HTTP request (GET /:code)
        |
        v
    redirectHandler
        |
        |-- looks up original URL in PostgreSQL
        |-- sends short_code to clickEvents channel (non-blocking)
        |-- immediately returns 302 redirect  <-- fast path, never waits
        |
        +-- (separately, in background) clickWorker goroutine
                |
                v
            consumes from clickEvents channel
                |
                v
            writes to click_events table in PostgreSQL  <-- slow path, async

This means a burst of clicks never slows down redirects, even under load - the channel buffers events (capacity 100) and the worker processes them at its own pace.

## Why Go

This project exists specifically to demonstrate Go's concurrency model - goroutines (lightweight, cheap threads) and channels (safe communication between them) - rather than duplicating the CRUD-API pattern already shown in the Python and TypeScript capstones. Go is the language of choice at companies building high-throughput infrastructure (Docker, Kubernetes, and most cloud-native tooling are written in Go) specifically because of this concurrency model.

## Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Service status |
| `/shorten` | POST | Create a short URL (`{"url": "..."}`) |
| `/:code` | GET | Redirect to the original URL; records a click asynchronously |
| `/stats?code=X` | GET | Get the click count for a short code |

## Running Locally

    createdb mydb  # if not already created
    psql mydb -c "CREATE TABLE IF NOT EXISTS short_urls (id SERIAL PRIMARY KEY, short_code VARCHAR(10) UNIQUE NOT NULL, original_url TEXT NOT NULL, created_at TIMESTAMP DEFAULT NOW());"
    psql mydb -c "CREATE TABLE IF NOT EXISTS click_events (id SERIAL PRIMARY KEY, short_code VARCHAR(10) NOT NULL, clicked_at TIMESTAMP DEFAULT NOW());"

    go build -o url-shortener main.go
    ./url-shortener

## Running Tests

    go test -v ./...

## Related Repositories

- [python-learning-journey](https://github.com/abdallasaidbigpartner-web/python-learning-journey)
- [typescript-learning-journey](https://github.com/abdallasaidbigpartner-web/typescript-learning-journey)
- [sql-learning-journey](https://github.com/abdallasaidbigpartner-web/sql-learning-journey)
- [ai-study-assistant](https://github.com/abdallasaidbigpartner-web/ai-study-assistant)
- [task-manager-api](https://github.com/abdallasaidbigpartner-web/task-manager-api)
- [ecommerce-database](https://github.com/abdallasaidbigpartner-web/ecommerce-database)
- [study-assistant-frontend](https://github.com/abdallasaidbigpartner-web/study-assistant-frontend)
