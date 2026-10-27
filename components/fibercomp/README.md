# Fiber Component `v0.3.3`

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
| `FIBER_ADDR` | **required** | Listen address (e.g. `:8080`) |
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
| `FIBER_OTEL_SKIP_PATHS` | `/health,/healthz` | Paths excluded from tracing |

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

## OpenTelemetry Support

Instrumentation is active only when `otelcomp.Otel` is registered — without it every provider is no-op. Paths listed in `FIBER_OTEL_SKIP_PATHS` are never traced.

### Trace

One server span per HTTP request, created by the [gofiber otel middleware](https://github.com/gofiber/contrib).

| Span | Kind | What it covers |
|------|------|----------------|
| `<METHOD> <route>` | `server` | One per request. The parent is taken from the incoming `traceparent` header |

| Attribute | Example |
|-----------|---------|
| `http.request.method` | `GET` |
| `url.path`, `url.query`, `url.scheme` | `/users`, `page=2`, `https` |
| `http.route` | `/users/:id` |
| `http.response.status_code` | `200` |
| `http.response.body.size` | `2` |
| `user_agent.original` | `curl/8.23.0` |
| `client.address` | `10.0.0.7` |
| `network.protocol.version` | `1.1` |
| `server.address`, `server.port` | `:3000` |
| `error.type` | `panic` |

Outgoing requests made with `http.DefaultClient` are traced as client spans by `otelcomp`.

### Metric

Recorded by the same middleware.

| Metric | Type | Unit |
|--------|------|------|
| `http.server.request.duration` | histogram | `s` |
| `http.server.active_requests` | up/down counter | `{request}` |
| `http.server.request.body.size` | histogram | `By` |
| `http.server.response.body.size` | histogram | `By` |

All four carry `http.request.method`, `http.route`, `http.response.status_code`, `server.address`, `server.port` and `url.scheme`.
