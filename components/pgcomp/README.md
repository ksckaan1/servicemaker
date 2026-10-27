# Postgres Component `v0.3.3`

PostgreSQL client component for ServiceMaker, using [pgx/v5](https://github.com/jackc/pgx) connection pool.

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/pgcomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/pgcomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    pg := sm.Register[pgcomp.Postgres]()

    var version string
    err := pg.DB().QueryRow(context.Background(), "SELECT version()").Scan(&version)
    if err != nil {
        log.Fatal(err)
    }
    log.Println(version)

    log.Fatal(sm.Run())
}
```

```bash
PG_DB_URL=postgres://user:pass@localhost:5432/mydb go run main.go
```

## Configuration

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `PG_DB_URL` | **required** | PostgreSQL connection URL (e.g. `postgres://user:pass@localhost:5432/dbname`) |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — creates the connection pool via `pgxpool.New()` |
| `Runner` | No — PostgreSQL has no blocking process |
| `Closer` | Yes — closes the connection pool |
| `HealthChecker` | Yes — pings the database |

## Methods

### `DB() *pgxpool.Pool`

Returns the underlying `pgxpool.Pool` for executing queries.

```go
pg := sm.Register[pgcomp.Postgres]()

rows, err := pg.DB().Query(ctx, "SELECT id, name FROM users")
```

## Lifecycle

1. **Init** — `pgxpool.New()` creates a connection pool using the provided `PG_DB_URL`.
2. **Run** — not implemented (returns immediately).
3. **Close** — `pgxpool.Pool.Close()` shuts down all connections in the pool.
4. **HealthCheck** — `pgxpool.Pool.Ping()` verifies the connection is alive.

## OpenTelemetry Support

Instrumentation is active only when `otelcomp.Otel` is registered — without it every provider is no-op.

### Trace

Provided by `github.com/exaring/otelpgx`, installed as the pgx `QueryTracer`.

| Span | Kind | What it covers |
|------|------|----------------|
| `SELECT`, `INSERT`, `UPDATE`, … | `client` | One per query, named after the SQL operation |
| `pool.acquire` | `client` | Waiting for a connection from the pool |
| `connect` | `client` | Creating a new pooled connection |
| `copy_from <table>` | `client` | One per `COPY FROM` |
| `batch start` | `client` | One per batch |

| Attribute | Example |
|-----------|---------|
| `db.system.name` | `postgresql` |
| `db.namespace` | `probe` |
| `db.query.text` | `SELECT 1` |
| `db.operation.name` | `SELECT` |
| `server.address`, `server.port` | `localhost`, `5432` |
| `user.name` | `postgres` |
| `pgx.rows_affected` | `1` |
| `pgx.prepare_stmt.name` | statement cache key, set on query spans that use a prepared statement |
| `error.type`, `exception.message` | on failure |

**Queries are only traced when the caller is already inside a recording span** — otelpgx returns early otherwise. Calls made inside a fiber/graphql request are covered; bare calls from `main` are not.

### Metric

| Metric | Type | Unit |
|--------|------|------|
| `db.client.operation.duration` | histogram | `s` |
| `pgxpool.acquire_duration` | histogram | `s` |
| `pgxpool.acquires` | counter | `{acquire}` |
| `pgxpool.canceled_acquires` | counter | `{acquire}` |
| `pgxpool.empty_acquire` | counter | `{acquire}` |
| `pgxpool.empty_acquire_wait_time` | histogram | `s` |
| `pgxpool.new_connections` | counter | `{connection}` |
| `pgxpool.max_connections` | up/down counter | `{connection}` |
| `pgxpool.total_connections` | up/down counter | `{connection}` |
| `pgxpool.idle_connections` | up/down counter | `{connection}` |
| `pgxpool.acquired_connections` | up/down counter | `{connection}` |
| `pgxpool.constructing_connections` | up/down counter | `{connection}` |
| `pgxpool.max_idle_destroys` | counter | `{connection}` |
| `pgxpool.max_lifetime_destroys` | counter | `{connection}` |

Pool metrics come from `otelpgx.RecordStats` and are always recorded; `db.client.operation.duration` comes from the tracer. Attributes include `db.system.name`, `db.client.connection.pool.name`, `db.operation.name` and `db.namespace`.
