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

## Pre-built Components

ServiceMaker ships with ready-to-use components. Import them and register directly — no extra setup needed:

| Package | Description |
|---------|-------------|
| `github.com/ksckaan1/servicemaker/fiber` | HTTP router based on [gofiber/fiber/v3](https://gofiber.io) |
| `github.com/ksckaan1/servicemaker/redis` | Redis client |

```go
sm.Register[fiberComp.Fiber]()
sm.Register[redisComp.Redis]()
```

Each pre-built component is a standalone Go module with its own `go.mod`. Import only the ones you need.

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    fiberComp "github.com/ksckaan1/servicemaker/fiber"
)

func main() {
    sm := servicemaker.New(context.Background())

    sm.Register[fiberComp.Fiber]()

    fiber := sm.Get[fiberComp.Fiber]()
    fiber.Router().Get("/", func(c fiber.Ctx) error {
        return c.SendString("Hello")
    })

    err := sm.Run()
    if err != nil {
        log.Fatal(err)
    }
}
```

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
