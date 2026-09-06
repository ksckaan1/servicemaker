# gRPC Component

gRPC server component for ServiceMaker, using [google.golang.org/grpc](https://github.com/grpc/grpc-go). Includes built-in logging interceptors.

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/grpccomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/grpccomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    grpc := sm.Register[grpccomp.GRPC]()

    pb.RegisterMyServiceServer(grpc.Server(), &myService{})

    log.Fatal(sm.Run())
}
```

```bash
GRPC_ADDR=:50051 go run main.go
```

## Configuration

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `GRPC_ADDR` | — | Listen address (e.g. `:50051`) |
| `GRPC_REFLECT` | `true` | Enable gRPC server reflection |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — creates the gRPC server with logging interceptors |
| `Runner` | Yes — starts serving on `GRPC_ADDR` |
| `Closer` | Yes — graceful stop |
| `HealthChecker` | No |

## Methods

### `Server() *grpc.Server`

Returns the underlying gRPC server for service registration.

### `AddUnaryServerInterceptor(interceptor grpc.UnaryServerInterceptor)`

Adds a unary server interceptor to the chain. Rebuilds the server with the updated interceptor chain.

### `AddStreamServerInterceptor(interceptor grpc.StreamServerInterceptor)`

Adds a stream server interceptor to the chain. Rebuilds the server with the updated interceptor chain.

## Built-in Interceptors

A logging interceptor is automatically added to both unary and stream chains. It logs the method name, duration, and any errors for each request.

## Lifecycle

1. **Init** — creates the gRPC server with logging interceptors.
2. **Run** — optionally registers reflection, then starts serving on `GRPC_ADDR`. Blocks until the context is cancelled.
3. **Close** — `GracefulStop()` stops the server gracefully.
