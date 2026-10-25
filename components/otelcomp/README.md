# OTel Component

OpenTelemetry tracing component for ServiceMaker. Sets up the global tracer provider with an OTLP/HTTP exporter, so the instrumentation built into other components (fibercomp, graphqlcomp, pgcomp, rediscomp, grpccomp) starts producing spans.

It also wraps `http.DefaultTransport` with `otelhttp`, so outgoing requests made with `http.DefaultClient` (or any client without its own `Transport`) are traced automatically and carry the W3C `traceparent` header.

Tracing is opt-in: if you never register this component, no provider is configured and all instrumentation stays no-op.

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/otelcomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/fibercomp"
    "github.com/ksckaan1/servicemaker/components/otelcomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    f := sm.Register[fibercomp.Fiber]()

    // Register last: its Close flushes buffered spans after the other
    // components have closed and ended their spans.
    sm.Register[otelcomp.Otel]()

    log.Fatal(sm.Run())
}
```

```bash
OTEL_SERVICE_NAME=my-service \
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318 \
go run main.go
```

## Configuration

Configuration is read by the OpenTelemetry SDK itself, using the standard environment variables:

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_SERVICE_NAME` | `unknown_service` | `service.name` resource attribute |
| `OTEL_RESOURCE_ATTRIBUTES` | — | Additional resource attributes (`key=value,key=value`) |
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `http://localhost:4318` | OTLP/HTTP collector endpoint |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | — | Traces-specific endpoint, overrides the generic one |
| `OTEL_EXPORTER_OTLP_HEADERS` | — | Headers sent to the collector |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — configures the global tracer provider |
| `Runner` | No |
| `Closer` | Yes — flushes buffered spans (5s timeout) |
| `HealthChecker` | No |

## Lifecycle

1. **Init** — creates the OTLP/HTTP exporter, builds the resource from the standard env variables, installs the global tracer provider and W3C trace-context propagator, then wraps `http.DefaultTransport` for outgoing request tracing.
2. **Close** — shuts down the tracer provider so buffered spans are exported. Uses a fresh 5 second context because the lifecycle context is already canceled at this point.
