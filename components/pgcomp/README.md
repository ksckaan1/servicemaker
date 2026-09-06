# Postgres Component

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
| `PG_DB_URL` | — | PostgreSQL connection URL (e.g. `postgres://user:pass@localhost:5432/dbname`) |

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
