# Redis Component `v0.4.1`

Redis client component for ServiceMaker, using [go-redis/v9](https://github.com/redis/go-redis). Provides per-DB connection pooling with lazy initialization.

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/rediscomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/rediscomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    rdb := sm.Register[rediscomp.Redis]()

    ctx := context.Background()
    rdb.DB(0).Set(ctx, "key", "value", 0)

    val, err := rdb.DB(0).Get(ctx, "key").Result()
    if err != nil {
        log.Fatal(err)
    }
    log.Println(val)

    log.Fatal(sm.Run())
}
```

## Configuration

All configuration is read from environment variables.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `REDIS_ADDR` | **required** | Redis server address (e.g. `localhost:6379`) |
| `REDIS_USER` | — | Username |
| `REDIS_PASS` | — | Password |
| `REDIS_CLIENT_NAME` | — | Client name sent via `CLIENT SETNAME` |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — initializes the connection pool map |
| `Runner` | No — Redis has no blocking process |
| `Closer` | Yes — closes all open connections |
| `HealthChecker` | Yes — pings all open connections |

## Methods

### `DB(db int) *redis.Client`

Returns a `*redis.Client` for the given DB number. Connections are created lazily on first access and reused thereafter.

```go
rdb := sm.Register[rediscomp.Redis]()

// Use DB 0
rdb.DB(0).Set(ctx, "key", "value", 0)

// Use DB 3
rdb.DB(3).Set(ctx, "key", "value", 0)
```

## Lifecycle

1. **Init** — creates an empty connection pool map. No connections are made yet.
2. **Run** — not implemented (returns immediately).
3. **Close** — iterates over all open connections and closes them. Logs errors for individual connection failures but continues closing the rest.
4. **HealthCheck** — sends `PING` to all open connections. Returns an error if any connection fails.

## OpenTelemetry Support

Instrumentation is active only when `otelcomp.Otel` is registered — without it every provider is no-op. Every connection handed out by `DB(db)` is instrumented, and each span and metric carries the extra `db_num` attribute so databases can be told apart.

### Trace

| Span | Kind | What it covers |
|------|------|----------------|
| `set`, `get`, `hgetall`, … | `client` | One per command, named after the Redis command |
| `redis.pipeline <summary>` | `client` | One per pipeline |
| `redis.dial` | `client` | Connection establishment |

| Attribute | Example |
|-----------|---------|
| `db.system` | `redis` |
| `db.connection_string` | `redis://localhost:6379` |
| `db.statement` | `SET probe 1` (truncated, may be disabled by redisotel options) |
| `db.redis.num_cmd` | `3` on pipeline spans |
| `db_num` | `0` — added by this component |
| `server.address`, `server.port` | `localhost`, `6379` |
| `code.function`, `code.filepath`, `code.lineno` | call site of the command |
| `error.type`, `exception.message` | on failure |

### Metric

Connection pool metrics from `redisotel.InstrumentMetrics`, using the OTel database semantic conventions.

| Metric | Type | Unit |
|--------|------|------|
| `db.client.connections.usage` | up/down counter | `{connection}` |
| `db.client.connections.idle.max` | up/down counter | `{connection}` |
| `db.client.connections.idle.min` | up/down counter | `{connection}` |
| `db.client.connections.max` | up/down counter | `{connection}` |
| `db.client.connections.waits` | counter | `{connection}` |
| `db.client.connections.waits_duration` | histogram | `s` |
| `db.client.connections.timeouts` | counter | `{timeout}` |
| `db.client.connections.hits` | counter | `{connection}` |
| `db.client.connections.misses` | counter | `{connection}` |
| `db.client.connections.create_time` | histogram | `s` |
| `db.client.connections.use_time` | histogram | `s` |

Attributes: `db.system`, `pool.name` (the Redis address) and `db_num`.
