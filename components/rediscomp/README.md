# Redis Component

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
| `REDIS_ADDR` | — | Redis server address (e.g. `localhost:6379`) |
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
