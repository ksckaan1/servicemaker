# Fiber Component

HTTP router component for ServiceMaker, based on [gofiber/fiber/v3](https://gofiber.io).

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/fibercomp
```

## Usage

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

## Configuration

All configuration is read from environment variables.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `FIBER_ADDR` | `:8080` | Listen address |
| `FIBER_BODY_LIMIT` | `0` | Max body size in bytes (`0` = unlimited) |
| `FIBER_MAX_RANGES` | `0` | Max number of range requests (`0` = unlimited) |
| `FIBER_CONCURRENCY` | `0` | Max concurrent connections (`0` = unlimited) |
| `FIBER_STREAM_REQUEST_BODY` | `false` | Stream request body instead of buffering |
| `FIBER_DISABLE_PRE_PARSE_MULTIPART_FORM` | `false` | Disable automatic multipart form pre-parsing |
| `FIBER_DISABLE_STARTUP_MESSAGE` | `true` | Disable the startup log message |
| `FIBER_USE_RECOVER_MW` | `true` | Enable panic recovery middleware |
| `FIBER_CORS_ALLOWED_ORIGINS` | — | CORS allowed origins (comma-separated). If set, CORS middleware is enabled |
| `FIBER_CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,PATCH,OPTIONS` | CORS allowed methods |
| `FIBER_CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization,Cookie` | CORS allowed headers |
| `FIBER_CORS_EXPOSE_HEADERS` | `Content-Length,Content-Type,Set-Cookie` | CORS exposed headers |
| `FIBER_CORS_MAX_AGE` | `86400` | CORS max age in seconds |
| `FIBER_CORS_ALLOW_CREDENTIALS` | `false` | Allow credentials in CORS |
| `FIBER_CORS_DISABLE_VALUE_REDUCTION` | `false` | Disable CORS value redaction |
| `FIBER_CORS_ALLOW_PRIVATE_NETWORK` | `false` | Allow private network access |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — creates the Fiber app with the provided config |
| `Runner` | Yes — starts listening on `FIBER_ADDR` |
| `Closer` | Yes — graceful shutdown |
| `HealthChecker` | No |

## Methods

### `Router() *fiber.App`

Returns the underlying Fiber app for route registration.

## Lifecycle

1. **Init** — `fiber.New()` is called with config derived from env vars.
2. **Run** — `fiber.Listen()` starts the server. Blocks until context is cancelled.
3. **Close** — `fiber.Shutdown()` performs a graceful shutdown.
