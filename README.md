# Task Management API

A RESTful API for creating and managing tasks with priority and status tracking, built with Go, Gin, PostgreSQL, and Redis.

## Tech Stack

| Layer       | Technology                        |
|-------------|-----------------------------------|
| HTTP        | [Gin](https://github.com/gin-gonic/gin) |
| Database    | PostgreSQL via [pgx v5](https://github.com/jackc/pgx) |
| Cache       | Redis via [go-redis v9](https://github.com/redis/go-redis) |
| Config      | Environment variables via [konf](https://github.com/nil-go/konf) |
| Logging     | [zap](https://github.com/uber-go/zap) with per-request ID |
| Docs        | [swaggo/swag](https://github.com/swaggo/swag) |

## Architecture

```
cmd/
  main.go                  ← entry point, wires everything
internal/
  config/                  ← configuration loading (env vars + defaults)
  delivery/
    handlers/              ← HTTP handlers (Gin)
    middleware/            ← request logger with request_id, recovery
    router/                ← route registration + Swagger UI
    dto/                   ← request/response structs
  migration/               ← embedded SQL migrations, run on startup
  models/                  ← domain models (Task, Priority, Status)
  ports/                   ← interfaces (Repository, UseCase, Handler, Cache)
  repository/              ← PostgreSQL implementation with row-level locking
  usecase/                 ← business logic
  cache/                   ← Redis implementation
  utils/
    logutil/               ← zap logger factory + context helpers
    errorutil/             ← typed HTTP errors
docs/                      ← generated Swagger files (do not edit)
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Redis

### Configuration

All configuration is via environment variables prefixed with `APP_`:

| Variable              | Default                                                   | Description          |
|-----------------------|-----------------------------------------------------------|----------------------|
| `APP_APP_ENV`         | `development`                                             | `development` or `production` |
| `APP_SERVER_HOST`     | `localhost`                                               | HTTP listen address  |
| `APP_SERVER_PORT`     | `8080`                                                    | HTTP listen port     |
| `APP_DATABASE_DSN`    | `postgres://localhost:5432/taskmanagement?sslmode=disable` | PostgreSQL DSN       |
| `APP_REDIS_HOST`      | `localhost`                                               | Redis host           |
| `APP_REDIS_PORT`      | `6379`                                                    | Redis port           |
| `APP_REDIS_PASSWORD`  | _(empty)_                                                 | Redis password       |
| `APP_REDIS_DB`        | `0`                                                       | Redis database index |

### Run

```bash
go run ./cmd/main.go
```

Migrations run automatically on startup — no separate migration step needed.

### Swagger UI

Once the server is running, open:

```
http://localhost:8080/swagger/index.html
```

To regenerate docs after changing handler annotations:

```bash
swag init -g cmd/main.go --output docs
```

## API Reference

Base path: `/api/v1`

### Tasks

| Method   | Path                  | Description              |
|----------|-----------------------|--------------------------|
| `GET`    | `/tasks`              | List tasks (paginated)   |
| `GET`    | `/tasks/:id`          | Get task by ID           |
| `POST`   | `/tasks`              | Create task              |
| `PUT`    | `/tasks/:id`          | Update task              |
| `DELETE` | `/tasks/:id`          | Delete task              |
| `PATCH`  | `/tasks/:id/status`   | Update task status       |

#### Query parameters — `GET /tasks`

| Parameter | Type   | Default | Description                         |
|-----------|--------|---------|-------------------------------------|
| `offset`  | int    | `0`     | Pagination offset                   |
| `limit`   | int    | `20`    | Page size                           |
| `search`  | string | —       | Filter by title or description      |

#### Priority values

`low` · `medium` · `high`

#### Status values

`todo` · `in_progress` · `done`

### Request tracing

Every response includes an `X-Request-ID` header. Pass the same header in a request to propagate your own ID end-to-end through the logs.

## Concurrency

`UpdateTask` and `DeleteTask` use a PostgreSQL `SELECT ... FOR UPDATE` inside a transaction to acquire a row-level lock before any modification, preventing lost updates under concurrent requests.
