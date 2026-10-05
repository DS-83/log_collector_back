# log_collector_back

[![Go](https://img.shields.io/github/go-mod/go-version/DS-83/log_collector_back)](https://go.dev)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-18%2B-336791?logo=postgresql&logoColor=white)](https://www.postgresql.org)
[![CI](https://github.com/DS-83/log_collector_back/actions/workflows/ci.yml/badge.svg)](https://github.com/DS-83/log_collector_back/actions/workflows/ci.yml)
[![License](https://img.shields.io/github/license/DS-83/log_collector_back?style=flat)](LICENSE)

A small, purpose-built HTTP event collector written in Go. It runs as a dedicated microservice next to a main web application, so application events and logs live in their own database instead of being mixed into the main one.

It is intentionally **not** a universal logging platform: it accepts structured events from one trusted application, stores them in PostgreSQL, and exposes a read API for an admin UI.

> **Status:** early development. The core ingest and read paths work; see the [roadmap](#roadmap) for what is planned.

## Features

- Event ingestion over HTTP, authenticated with API keys
- Idempotent writes: clients generate `event_id`, retries never create duplicates
- Event source is taken from the API key, not from the request body, so a client cannot write events on behalf of another source
- Paginated, sortable, searchable event list with a time range (7 days by default)
- Event details by ID
- Session-based authentication (HttpOnly cookie) for the admin UI
- Events table partitioned by time for cheap retention
- Graceful shutdown, per-request context timeout

## How it works

```mermaid
flowchart LR
    App["Main web app"] -- "POST /api/v1/events/create<br/>X-Api-Key" --> C["log_collect"]
    UI["Admin UI"] -- "GET /events/*<br/>session cookie" --> C
    C --> DB[("PostgreSQL<br/>separate database")]
```

There are two independent authentication mechanisms, because they serve different callers:

| | Ingest | Admin UI |
|---|---|---|
| Caller | a service | a person |
| Credential | `X-Api-Key` header | session cookie |
| Stored in DB | SHA-256 hash of the key | bcrypt password hash, SHA-256 hash of the session token |
| Routes | `/api/v1/events/*` | `/events/*` |

Raw API keys and session tokens are never stored. Keys are 32 random bytes encoded as base64url and shown only once, at creation.

## Tech stack

- Go, [Gin](https://github.com/gin-gonic/gin)
- PostgreSQL 18+ (uses the native `uuidv7()` function) via [pgx](https://github.com/jackc/pgx) (`pgxpool`)
- [sqlc](https://sqlc.dev) for type-safe queries
- [goose](https://github.com/pressly/goose) for migrations
- [viper](https://github.com/spf13/viper) + [godotenv](https://github.com/joho/godotenv) for configuration
- bcrypt for password hashing

## Project structure

The project follows a hexagonal (ports and adapters) layout: business logic depends on interfaces, and infrastructure implements them.

sqlc generates one package per group of queries (for example `apikeysdb`, `usersdb`, `eventsdb`), so each repository only sees the queries for its own tables.

## Getting started

### Prerequisites

- Go (see `go.mod` for the version)
- PostgreSQL 18 or newer
- [goose](https://github.com/pressly/goose) for migrations
- [sqlc](https://sqlc.dev), only if you change SQL queries

### Configuration

Configuration comes from two sources:

- **`config/config.dev.yml`** and **`config/config.prod.yml`** for non-secret settings. The production file is used when `APP_ENV=production`.
- **Environment variables** (optionally from a `.env` file) for secrets.

Environment:

```env
# .env
DB_DSN=postgres://user:password@localhost:5432/log_collect?sslmode=disable
```

`config/config.dev.yml`:

```yaml
app:
  port: "8080"
context:
  timeout: 10s      # per-request timeout, also used for server read/write timeouts
session:
  ttl: 24h
cookie:
  name: session
  path: /
  domain: ""
  secure: false     # set to true behind HTTPS
  http_only: true
hasher:
  cost: 12          # bcrypt cost
```

### Database

Create an empty database and apply the migrations:

```bash
goose -dir internal/infrastructure/repo/pgsql/migrations postgres "$DB_DSN" up
```

### First user

There is no public sign-up. Generate a password hash with the helper and insert the user into the `users` table yourself:

```bash
go run ./cmd/hashpass
```

The helper uses the same bcrypt cost as the server, taken from the config.

### API key

```bash
go run ./cmd/keygen
```

It prints a new key and its hash. Give the key to the client application and insert the hash into `api_keys` (`name`, `key_hash`, `source`). `key_hash` is a `bytea` column, so in SQL the hash is written as `'\x<hex>'`.

### Run

```bash
go run ./cmd/api
```
### Docker (example)

`docker-compose.example.yml` is a sample deployment behind Caddy, not a ready-made production setup. Replace the external network name `caddy_caddy-internet` with the name of your own Caddy network (list networks with `docker network ls`).

## API

### Send an event

`POST /api/v1/events/create`, authenticated with the `X-Api-Key` header.

```bash
curl -X POST http://localhost:8080/api/v1/events/create \
  -H "X-Api-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "event_id": "0197a1c2-7b3e-7c3a-9d52-3f4a1b2c3d4e",
    "occurred_at": "2026-10-05T12:00:00Z",
    "event_type": "auth.login",
    "level": "info",
    "actor_id": "user-42",
    "request_id": "req-abc123",
    "payload": { "ip": "203.0.113.7" }
  }'
```

| Field | Type | Required | Notes |
|---|---|---|---|
| `event_id` | UUID | yes | generated by the client; repeated IDs are ignored |
| `occurred_at` | RFC 3339 timestamp | yes | when the event happened in the client app |
| `event_type` | string | yes | for example `auth.login` |
| `level` | string | yes | `debug`, `info`, `warn` or `error` |
| `actor_id` | string | no | user ID in the main application |
| `request_id` | string | no | correlation ID of the originating request |
| `payload` | JSON object | no | arbitrary data, stored as `jsonb` |

Responds with `201 Created` and an empty body. The event's `source` is set from the API key.

### Sign in

`POST /userauth/signin` with `{"login": "...", "password": "..."}`. On success the server sets the session cookie.

### List events

`GET /events/list`, requires the session cookie.

| Query parameter | Description |
|---|---|
| `query` | substring search over event type and source |
| `from`, `to` | RFC 3339 time range; defaults to the last 7 days, at most 31 days |
| `limit` | page size, default 50, maximum 200 |
| `offset` | number of rows to skip |
| `sortField` | `occurred_at` (default), `type`, `source` or `level` |
| `order` | `asc` (default) or `desc` |

```bash
curl -b "session=<token>" \
  "http://localhost:8080/events/list?limit=20&sortField=occurred_at&order=desc"
```

Response:

```json
{ "events": [ { "id": 1, "event_id": "...", "occurred_at": "...", "source": "api", "event_type": "auth.login", "level": "info" } ], "count": 1 }
```

`count` is the total number of matching rows, for rendering pagination.

### Get one event

`GET /events/getone/:id`, requires the session cookie. Returns the full event, including `payload`, `actor_id`, `request_id` and `received_at`.

## Data model

| Table | Purpose |
|---|---|
| `events` | collected events, range-partitioned by `occurred_at`, with a default partition |
| `api_keys` | ingest credentials: key hash, source, revocation time |
| `users` | admin UI users |
| `sessions` | server-side sessions: token hash, user, expiry |
| `event_types` | registry of event types (reserved for enabling/disabling types) |

Event levels are stored as numbers (`debug` 10, `info` 20, `warn` 30, `error` 40) so intermediate levels can be added later.

## Design notes

- **Separate database.** Logs have a different write profile and retention than business data, so they never share a database with the main application.
- **No foreign keys from `events`.** `actor_id` and `request_id` are plain text; the collector stays decoupled from the main application's schema.
- **Idempotency.** `UNIQUE (occurred_at, event_id)` plus `ON CONFLICT DO NOTHING`. The unique key includes `occurred_at` because it is the partition key.
- **Stable pagination.** The list is always ordered by the chosen field plus `id` as a tie-breaker, so equal values never reshuffle between pages.
- **Bounded queries.** A time range is always applied and its length and the page size are capped, so a single request cannot scan the whole table.
- **Safe dynamic sorting.** Sort columns come from an allow-list; user input is never concatenated into SQL.

## Development

Regenerate query code after changing anything in `queries/`:

```bash
cd internal/infrastructure/repo/pgsql
sqlc generate
```

Create a migration and run the tests:

```bash
goose -dir internal/infrastructure/repo/pgsql/migrations create <name> sql
go test ./...
```

Already-applied migrations should not be edited; add a new one instead.

## Roadmap

- [ ] Batch ingestion endpoint
- [ ] Automatic creation of time partitions and retention by dropping old ones
- [ ] Event type registry: allow-list, enabling and disabling types at runtime
- [ ] Endpoint to manage API keys from the admin UI
- [ ] Web UI