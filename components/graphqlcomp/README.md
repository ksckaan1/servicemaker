# GraphQL Component

GraphQL server component for ServiceMaker, based on [gqlgen](https://github.com/99designs/gqlgen) with a [gofiber/fiber/v3](https://gofiber.io) HTTP transport.

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/graphqlcomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/99designs/gqlgen/graphql/handler"
    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/graphqlcomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    gql := sm.Register[graphqlcomp.GraphQL]()

    srv := handler.NewDefaultServer(executableSchema)
    gql.RegisterServer("/graphql", srv)

    log.Fatal(sm.Run())
}
```

```bash
GQL_ADDR=:8080 GQL_PLAYGROUND=true go run main.go
```

## Configuration

All configuration is read from environment variables.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `GQL_ADDR` | `:8080` | Listen address |
| `GQL_BODY_LIMIT` | `0` | Max body size in bytes (`0` = unlimited) |
| `GQL_MAX_RANGES` | `0` | Max number of range requests (`0` = unlimited) |
| `GQL_CONCURRENCY` | `0` | Max concurrent connections (`0` = unlimited) |
| `GQL_STREAM_REQUEST_BODY` | `false` | Stream request body instead of buffering |
| `GQL_DISABLE_PRE_PARSE_MULTIPART_FORM` | `false` | Disable automatic multipart form pre-parsing |
| `GQL_DISABLE_STARTUP_MESSAGE` | `true` | Disable the startup log message |
| `GQL_USE_RECOVER_MW` | `true` | Enable panic recovery middleware |
| `GQL_INTROSPECTION` | `false` | Enable GraphQL introspection |
| `GQL_PLAYGROUND` | `false` | Enable GraphQL playground at `{endpoint}-playground` |
| `GQL_CORS_ALLOWED_ORIGINS` | — | CORS allowed origins (comma-separated). If set, CORS middleware is enabled |
| `GQL_CORS_ALLOWED_METHODS` | `GET,POST,PUT,DELETE,PATCH,OPTIONS` | CORS allowed methods |
| `GQL_CORS_ALLOWED_HEADERS` | `Origin,Content-Type,Accept,Authorization,Cookie` | CORS allowed headers |
| `GQL_CORS_EXPOSE_HEADERS` | `Content-Length,Content-Type,Set-Cookie` | CORS exposed headers |
| `GQL_CORS_MAX_AGE` | `86400` | CORS max age in seconds |
| `GQL_CORS_ALLOW_CREDENTIALS` | `false` | Allow credentials in CORS |
| `GQL_CORS_ALLOW_PRIVATE_NETWORK` | `false` | Allow private network access |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — creates the Fiber app with the provided config |
| `Runner` | Yes — starts listening on `GQL_ADDR` |
| `Closer` | Yes — graceful shutdown |
| `HealthChecker` | No |

## Methods

### `RegisterServer(endpoint string, server *handler.Server)`

Registers a gqlgen handler server at the given endpoint. Optionally enables introspection and playground based on config.

```go
gql := sm.Register[graphqlcomp.GraphQL]()

srv := handler.NewDefaultServer(executableSchema)
gql.RegisterServer("/graphql", srv)
```

## Lifecycle

1. **Init** — `fiber.New()` is called with config derived from env vars.
2. **Run** — `fiber.Listen()` starts the server. Blocks until context is cancelled.
3. **Close** — `fiber.Shutdown()` performs a graceful shutdown.
