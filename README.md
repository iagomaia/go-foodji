# go-foodji

A Go REST API for managing product voting sessions. Users create a session and cast a like or dislike vote for any product within that session — one vote per product per session, updated atomically.

## Requirements

- [Go 1.22+](https://go.dev/dl/)
- [Docker](https://www.docker.com/) (for local MongoDB)

## Getting started

1. Copy the example env file and adjust if needed:
   ```bash
   cp .env.example .env
   ```

2. Start MongoDB:
   ```bash
   make docker-up
   ```

3. Run the API:
   ```bash
   make start
   ```

The server will be available at `http://localhost:8080`.

## Environment variables

| Variable           | Default                      | Description                  |
|--------------------|------------------------------|------------------------------|
| `APP_PORT`         | `8080`                       | HTTP port the server listens on |
| `APP_ENV`          | `development`                | Execution environment (`development` / `production`) |
| `MONGODB_URI`      | `mongodb://localhost:27017`  | MongoDB connection URI       |
| `MONGODB_DATABASE` | `foodji`                     | MongoDB database name        |

## Running tests

```bash
make test
```

Tests run with the race detector enabled. All tests are unit tests — no database required.

## Swagger UI

**Live playground:** [https://foodji-tech-challenge-yg2vj.ondigitalocean.app/playground/index.html](https://foodji-tech-challenge-yg2vj.ondigitalocean.app/playground/index.html)

When running locally (`APP_ENV=development` or `APP_ENV=local`), the Swagger UI is also available at:

```
http://localhost:8080/playground/index.html
```

It is **not registered** in production (`APP_ENV=production`).

To regenerate the OpenAPI spec after changing handler annotations:

```bash
make swag
```

This rewrites `docs/swagger.yaml` and `docs/docs.go` from source annotations.

## Makefile targets

| Target             | Description                                      |
|--------------------|--------------------------------------------------|
| `make start`       | Run the API with `go run`                        |
| `make build`       | Compile binary to `bin/api`                      |
| `make test`        | Run all tests with race detection                |
| `make swag`        | Regenerate `docs/swagger.yaml` and `docs/docs.go` |
| `make docker-up`   | Start MongoDB via Docker Compose                 |
| `make docker-down` | Stop Docker Compose containers                   |

---

## Architecture

The project follows a clean layered architecture where each layer depends only on the layer below it via interfaces, making every component independently testable.

```
cmd/api/
└── main.go               # Wires dependencies and starts the server

internal/
├── config/               # Loads configuration from environment / .env
├── domain/               # Pure Go types: entities, inputs, errors — no dependencies
├── handler/              # HTTP layer: binds requests, calls services, writes responses
├── middleware/           # Cross-cutting HTTP concerns (request logging)
├── repository/           # Repository interfaces (contracts, no implementation)
│   └── mongo/            # MongoDB implementations of the repository interfaces
├── server/               # Gin router setup and route registration
└── service/              # Business logic; depends on repository interfaces only
    ├── session/
    └── vote/

pkg/
├── logger/               # Structured logger (slog-based)
└── telemetry/            # OpenTelemetry tracing setup (stdout exporter)

test/
└── mocks/                # Hand-written interface mocks for unit tests
```

### Dependency flow

```
handler → service → repository interface ← mongo implementation
```

Handlers know nothing about MongoDB. Services know nothing about HTTP. Mongo implementations know nothing about business rules.

### Key design decisions

- **Atomic upsert** — `PUT /votes` performs a single `UpdateOne` with `upsert=true` against MongoDB, using `$set` for mutable fields and `$setOnInsert` for immutable ones. No read-before-write.
- **Unique compound index** — A unique index on `(session_id, product_id)` is created at startup via `EnsureIndexes`, enforcing the one-vote-per-product-per-session invariant at the database level.
- **Session gate** — A vote can only be registered if the referenced session exists. The service validates this before calling the repository.

### Dependencies

| Package | Purpose |
|---------|---------|
| [gin-gonic/gin](https://github.com/gin-gonic/gin) | HTTP router and middleware |
| [spf13/viper](https://github.com/spf13/viper) | Configuration from env / .env files |
| [go.mongodb.org/mongo-driver/v2](https://github.com/mongodb/mongo-go-driver) | Official MongoDB driver |
| [go.opentelemetry.io/otel](https://opentelemetry.io/docs/languages/go/) | Distributed tracing (stdout exporter) |
| [otelgin](https://github.com/open-telemetry/opentelemetry-go-contrib/tree/main/instrumentation/github.com/gin-gonic/gin/otelgin) | OpenTelemetry middleware for Gin |
| [stretchr/testify](https://github.com/stretchr/testify) | Test assertions and helpers |

---

## API

Base path: `/api/v1`

### Health check

```
GET /health
```

**Response `200`**
```json
{ "status": "ok" }
```

---

### Sessions

#### Create a session

```
POST /api/v1/sessions
```

No request body required.

**Response `201`**
```json
{
  "id": "6642f1a2b3c4d5e6f7a8b9c0",
  "created_at": "2024-05-14T10:00:00Z",
  "updated_at": "2024-05-14T10:00:00Z"
}
```

---

### Votes

#### Upsert a vote

Registers or updates a vote for a product within a session. Only one vote per product per session is allowed — submitting again overwrites the previous vote type.

Returns `201` when the vote is newly created, `200` when an existing vote is updated.

```
PUT /api/v1/votes
```

**Request body**
```json
{
  "session_id": "6642f1a2b3c4d5e6f7a8b9c0",
  "product_id": "prod-abc-123",
  "vote_type": "like"
}
```

| Field        | Type   | Values              | Required |
|--------------|--------|---------------------|----------|
| `session_id` | string | valid session ID    | yes      |
| `product_id` | string | any string          | yes      |
| `vote_type`  | string | `"like"`, `"dislike"` | yes    |

**Response `201`** (new vote)
```json
{
  "id": "6642f2b3c4d5e6f7a8b9c0d1",
  "session_id": "6642f1a2b3c4d5e6f7a8b9c0",
  "product_id": "prod-abc-123",
  "vote_type": "like",
  "created_at": "2024-05-14T10:01:00Z",
  "updated_at": "2024-05-14T10:01:00Z"
}
```

**Response `200`** (updated vote)
```json
{
  "id": "6642f2b3c4d5e6f7a8b9c0d1",
  "session_id": "6642f1a2b3c4d5e6f7a8b9c0",
  "product_id": "prod-abc-123",
  "vote_type": "dislike",
  "created_at": "2024-05-14T10:01:00Z",
  "updated_at": "2024-05-14T10:05:00Z"
}
```

**Response `404`** — session not found

---

#### List votes for a session

```
GET /api/v1/votes/sessions/:session_id
```

**Response `200`**
```json
[
  {
    "id": "6642f2b3c4d5e6f7a8b9c0d1",
    "session_id": "6642f1a2b3c4d5e6f7a8b9c0",
    "product_id": "prod-abc-123",
    "vote_type": "like",
    "created_at": "2024-05-14T10:01:00Z",
    "updated_at": "2024-05-14T10:01:00Z"
  },
  {
    "id": "6642f2b3c4d5e6f7a8b9c0d2",
    "session_id": "6642f1a2b3c4d5e6f7a8b9c0",
    "product_id": "prod-xyz-456",
    "vote_type": "dislike",
    "created_at": "2024-05-14T10:02:00Z",
    "updated_at": "2024-05-14T10:02:00Z"
  }
]
```

---

#### Get vote report

Aggregates like and dislike counts across all sessions, with a like/dislike ratio per product.

```
GET /api/v1/votes/report
```

**Response `200`**
```json
[
  {
    "product_id": "prod-abc-123",
    "like_count": 42,
    "like_ratio": 0.875,
    "dislike_count": 6,
    "dislike_ratio": 0.125
  },
  {
    "product_id": "prod-xyz-456",
    "like_count": 10,
    "like_ratio": 0.5,
    "dislike_count": 10,
    "dislike_ratio": 0.5
  }
]
```
