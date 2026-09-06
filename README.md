# ServiceMaker

A Go library that orchestrates the lifecycle of application components (init, run, close, health check). Component configuration is read from environment variables.

## How it works

An application typically consists of multiple independent components: an HTTP server, a database connection, Redis, etc. Each component follows a similar lifecycle:

1. **Init** — Establish connections, prepare resources.
2. **Run** — Start listening (usually blocking).
3. **Close** — Tear down connections, release resources.
4. **HealthCheck** — Check health status.

ServiceMaker orchestrates this lifecycle from a single central point. You add components via `Register`, then start them all concurrently with `Run`. When a signal arrives (`SIGINT`/`SIGKILL`), all `Closer` implementations are called automatically in order.

## Interfaces

A component only implements the interfaces for the stages it participates in — it does not need to implement all of them:

```go
type Initializer interface {
    Init(ctx context.Context) error
}

type Runner interface {
    Run(ctx context.Context) error
}

type Closer interface {
    Close(ctx context.Context) error
}

type HealthChecker interface {
    HealthCheck(ctx context.Context) error
}
```

## Getting Started

ServiceMaker ships with pre-built components so you can start immediately without writing boilerplate.

### Quick Start

```bash
go get github.com/ksckaan1/servicemaker
go get github.com/ksckaan1/servicemaker/components/fibercomp
```

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/fibercomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    f := sm.Register[fibercomp.Fiber]()

    f.Router().Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    log.Fatal(sm.Run())
}
```

```bash
FIBER_ADDR=:3000 go run main.go
```

### Multiple Components

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/fibercomp"
    "github.com/ksckaan1/servicemaker/components/rediscomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    f := sm.Register[fibercomp.Fiber]()
    rdb := sm.Register[rediscomp.Redis]()

    rdb.DB(0).Set(context.Background(), "key", "value", 0)

    f.Router().Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello, World!")
    })

    log.Fatal(sm.Run())
}
```

```bash
FIBER_ADDR=:3000 REDIS_ADDR=localhost:6379 go run main.go
```

## Pre-built Components

Each component is a standalone Go module. Import only the ones you need.

| Package | Description | Docs |
|---------|-------------|------|
| `components/fibercomp` | HTTP router ([gofiber/fiber/v3](https://gofiber.io)) | [README](components/fibercomp/README.md) |
| `components/rediscomp` | Redis client ([go-redis/v9](https://github.com/redis/go-redis)) | [README](components/rediscomp/README.md) |
| `components/pgcomp` | PostgreSQL client ([pgx/v5](https://github.com/jackc/pgx)) | [README](components/pgcomp/README.md) |
| `components/rabbitmqcomp` | RabbitMQ client ([amqp091-go](https://github.com/rabbitmq/amqp091-go)) | [README](components/rabbitmqcomp/README.md) |

## Custom Components

To create a custom component, define a struct and implement one or more of the ServiceMaker interfaces. Configuration is parsed from environment variables using `env` struct tags.

### Minimal Component (Init only)

```go
package mycomponent

import "context"

type Database struct {
    DSN string `env:"MY_DB_DSN"`
    db  *sql.DB
}

func (d *Database) Init(ctx context.Context) error {
    var err error
    d.db, err = sql.Open("postgres", d.DSN)
    if err != nil {
        return fmt.Errorf("sql.Open: %w", err)
    }
    return d.db.PingContext(ctx)
}
```

```go
db := sm.Register[mycomponent.Database]()
```

### Full Lifecycle Component

```go
package mycomponent

import (
    "context"
    "fmt"
    "net/http"
)

type Server struct {
    Addr string `env:"SERVER_ADDR" envDefault:":9090"`
    srv  *http.Server
}

func (s *Server) Init(ctx context.Context) error {
    s.srv = &http.Server{Addr: s.Addr}
    return nil
}

func (s *Server) Run(ctx context.Context) error {
    <-ctx.Done()
    return nil
}

func (s *Server) Close(ctx context.Context) error {
    return s.srv.Shutdown(ctx)
}

func (s *Server) HealthCheck(ctx context.Context) error {
    resp, err := http.Get("http://" + s.Addr + "/health")
    if err != nil {
        return err
    }
    resp.Body.Close()
    if resp.StatusCode != http.StatusOK {
        return fmt.Errorf("unhealthy: status %d", resp.StatusCode)
    }
    return nil
}
```

### Config Tag Reference

ServiceMaker uses [caarlos0/env](https://github.com/caarlos0/env) for parsing. Common tags:

| Tag                  | Description               |
| -------------------- | ------------------------- |
| `env:"VAR_NAME"`     | Environment variable name |
| `envDefault:"value"` | Default value if unset    |

## Component State Machine

```
Register[T]()  → parse env config → Init(ctx) → (ready)
Run()          → Runner.Run(ctx) concurrently (errgroup)
SIGINT/SIGKILL → Close(ctx) in order → Run() returns
```

If a component does not implement the `Runner` interface, it is skipped during the `Run` phase. Same for `Closer` during shutdown.

## Installation

```bash
go get github.com/ksckaan1/servicemaker
```

## License

[MIT](LICENSE)
