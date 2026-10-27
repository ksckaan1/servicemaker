# OTel Component

OpenTelemetry observability component for ServiceMaker. Sets up the global tracer, logger and meter providers with OTLP/HTTP exporters, so the instrumentation built into other components (fibercomp, graphqlcomp, pgcomp, rediscomp, grpccomp) starts producing spans and metrics, and logs written through `github.com/ksckaan1/logger` are exported as OTLP log records.

It also wraps `http.DefaultTransport` with `otelhttp`, so outgoing requests made with `http.DefaultClient` (or any client without its own `Transport`) are traced automatically and carry the W3C `traceparent` header.

Observability is opt-in: if you never register this component, no provider is configured, so all instrumentation stays no-op and logs only go to stdout/file.

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

    // Register last: its Close flushes buffered spans and logs after the
    // other components have closed and ended their spans.
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

All configuration is read from environment variables by the OpenTelemetry Go SDK itself — this component does not parse them. The tables below list every variable the Go SDK actually honours in the tracing and logging paths, matching the [OpenTelemetry environment variable specification](https://opentelemetry.io/docs/specs/otel/configuration/sdk-environment-variables/) and the `otlptracehttp` / `otlploghttp` exporter docs for `go.opentelemetry.io/otel` v1.47.0.

### Resource

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_SERVICE_NAME` | `unknown_service` | Sets the `service.name` resource attribute. It takes precedence over a `service.name` given in `OTEL_RESOURCE_ATTRIBUTES`. |
| `OTEL_RESOURCE_ATTRIBUTES` | — | Extra resource attributes as comma-separated `key=value` pairs. Values are URL-encoded, e.g. `deployment.environment=prod,service.version=1.2.3`. |
| `OTEL_GO_X_RESOURCE` | `false` | Go-specific experimental flag. Set to `true` to let resource detectors add experimental semantic-convention attributes (e.g. the newer `service.instance.id`/`host.*` conventions). |

### OTLP/HTTP exporter

Trace, log and metric exporters are separate but share the same generic variables. Every generic `OTEL_EXPORTER_OTLP_*` variable has a signal-specific `OTEL_EXPORTER_OTLP_TRACES_*` / `OTEL_EXPORTER_OTLP_LOGS_*` / `OTEL_EXPORTER_OTLP_METRICS_*` counterpart. The signal-specific one always wins.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | `https://localhost:4318` | Collector base URL; `/v1/traces`, `/v1/logs` or `/v1/metrics` is appended. Must contain a scheme (`http` or `https`) and host, optionally a port and path. No query string or fragment. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT`<br>`OTEL_EXPORTER_OTLP_LOGS_ENDPOINT`<br>`OTEL_EXPORTER_OTLP_METRICS_ENDPOINT` | `https://localhost:4318/v1/{traces,logs,metrics}` | Full signal URL, used as-is (no path appending). |
| `OTEL_EXPORTER_OTLP_PROTOCOL`<br>`OTEL_EXPORTER_OTLP_TRACES_PROTOCOL`<br>`OTEL_EXPORTER_OTLP_LOGS_PROTOCOL`<br>`OTEL_EXPORTER_OTLP_METRICS_PROTOCOL` | `http/protobuf` | Payload encoding. Supported: `http/protobuf`, `http/json`. |
| `OTEL_EXPORTER_OTLP_TIMEOUT`<br>`OTEL_EXPORTER_OTLP_TRACES_TIMEOUT`<br>`OTEL_EXPORTER_OTLP_LOGS_TIMEOUT`<br>`OTEL_EXPORTER_OTLP_METRICS_TIMEOUT` | `10000` | Max time in milliseconds the exporter waits for a single batch export. |
| `OTEL_EXPORTER_OTLP_HEADERS`<br>`OTEL_EXPORTER_OTLP_TRACES_HEADERS`<br>`OTEL_EXPORTER_OTLP_LOGS_HEADERS`<br>`OTEL_EXPORTER_OTLP_METRICS_HEADERS` | — | Headers sent with every export request, e.g. `Authorization=Bearer%20token,X-Tenant=acme`. Format is W3C Baggage style (`key=value,key=value`) without the extra metadata. |
| `OTEL_EXPORTER_OTLP_COMPRESSION`<br>`OTEL_EXPORTER_OTLP_TRACES_COMPRESSION`<br>`OTEL_EXPORTER_OTLP_LOGS_COMPRESSION`<br>`OTEL_EXPORTER_OTLP_METRICS_COMPRESSION` | none | Request body compression. Supported: `gzip`. |
| `OTEL_EXPORTER_OTLP_INSECURE`<br>`OTEL_EXPORTER_OTLP_TRACES_INSECURE`<br>`OTEL_EXPORTER_OTLP_LOGS_INSECURE`<br>`OTEL_EXPORTER_OTLP_METRICS_INSECURE` | `false` | `true` disables TLS for the collector connection. Set this (or use an `http://` endpoint) when talking to a plain-HTTP collector such as a local `otel/opentelemetry-collector`. |
| `OTEL_EXPORTER_OTLP_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_TRACES_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_LOGS_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_METRICS_CERTIFICATE` | — | Path to a PEM file with trusted CA certificates used to verify the collector's TLS certificate. |
| `OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_TRACES_CLIENT_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_LOGS_CLIENT_CERTIFICATE`<br>`OTEL_EXPORTER_OTLP_METRICS_CLIENT_CERTIFICATE` | — | Path to a PEM client certificate/chain, for mTLS. |
| `OTEL_EXPORTER_OTLP_CLIENT_KEY`<br>`OTEL_EXPORTER_OTLP_TRACES_CLIENT_KEY`<br>`OTEL_EXPORTER_OTLP_LOGS_CLIENT_KEY`<br>`OTEL_EXPORTER_OTLP_METRICS_CLIENT_KEY` | — | Path to the PEM client private key, for mTLS. |

### Batch processors

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_BSP_SCHEDULE_DELAY` | `5000` | Trace export delay in milliseconds between two consecutive exports. |
| `OTEL_BSP_EXPORT_TIMEOUT` | `30000` | Max time in milliseconds a trace export may take before it is cancelled. |
| `OTEL_BSP_MAX_QUEUE_SIZE` | `2048` | Max number of finished spans buffered before new ones are dropped. |
| `OTEL_BSP_MAX_EXPORT_BATCH_SIZE` | `512` | Max number of spans sent per export. Must be ≤ `OTEL_BSP_MAX_QUEUE_SIZE`. |
| `OTEL_BLRP_SCHEDULE_DELAY` | `1000` | Log export delay in milliseconds between two consecutive exports. |
| `OTEL_BLRP_EXPORT_TIMEOUT` | `30000` | Max time in milliseconds a log export may take before it is cancelled. |
| `OTEL_BLRP_MAX_QUEUE_SIZE` | `2048` | Max number of log records buffered before new ones are dropped. |
| `OTEL_BLRP_MAX_EXPORT_BATCH_SIZE` | `512` | Max number of log records sent per export. Must be ≤ `OTEL_BLRP_MAX_QUEUE_SIZE`. |

### Sampling

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_TRACES_SAMPLER` | `parentbased_always_on` | Sampling strategy. Supported: `always_on`, `always_off`, `traceidratio`, `parentbased_always_on`, `parentbased_always_off`, `parentbased_traceidratio`. |
| `OTEL_TRACES_SAMPLER_ARG` | `1.0` | Argument for the sampler; the ratio in `[0.0, 1.0]` for `traceidratio` and `parentbased_traceidratio`. Ignored by the other samplers. |

### Span limits

The `OTEL_SPAN_*` variable wins over its generic `OTEL_*` fallback. Set a length limit to `-1` for unlimited.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_SPAN_ATTRIBUTE_VALUE_LENGTH_LIMIT`<br>`OTEL_ATTRIBUTE_VALUE_LENGTH_LIMIT` | `-1` (unlimited) | Max characters kept per attribute value; longer strings are truncated. |
| `OTEL_SPAN_ATTRIBUTE_COUNT_LIMIT`<br>`OTEL_ATTRIBUTE_COUNT_LIMIT` | `128` | Max number of attributes per span. |
| `OTEL_SPAN_EVENT_COUNT_LIMIT` | `128` | Max number of events per span. |
| `OTEL_EVENT_ATTRIBUTE_COUNT_LIMIT` | `128` | Max number of attributes per span event. |
| `OTEL_SPAN_LINK_COUNT_LIMIT` | `128` | Max number of links per span. |
| `OTEL_LINK_ATTRIBUTE_COUNT_LIMIT` | `128` | Max number of attributes per span link. |

### Metrics

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_METRIC_EXPORT_INTERVAL` | `60000` | How often the periodic reader collects and exports metrics, in milliseconds. |
| `OTEL_METRIC_EXPORT_TIMEOUT` | `30000` | Max time in milliseconds a metric export may take before it is cancelled. |
| `OTEL_EXPORTER_OTLP_METRICS_TEMPORALITY_PREFERENCE` | `cumulative` | Aggregation temporality per instrument kind. Supported: `cumulative`, `delta`, `lowmemory`. |
| `OTEL_EXPORTER_OTLP_METRICS_DEFAULT_HISTOGRAM_AGGREGATION` | `explicit_bucket_histogram` | Default histogram aggregation. Supported: `explicit_bucket_histogram`, `base2_exponential_bucket_histogram`. |
| `OTEL_METRICS_EXEMPLAR_FILTER` | `trace_based` | Which measurements keep exemplars. Supported: `always_on`, `always_off`, `trace_based`. |

### Log record limits

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `OTEL_LOGRECORD_ATTRIBUTE_COUNT_LIMIT` | `128` | Max number of attributes per log record. |
| `OTEL_LOGRECORD_ATTRIBUTE_VALUE_LENGTH_LIMIT` | `-1` (unlimited) | Max characters kept per log record attribute value. |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — configures the global tracer, logger and meter providers |
| `Runner` | No |
| `Closer` | Yes — flushes buffered spans, log records and metrics (5s timeout) |
| `HealthChecker` | No |

## Lifecycle

1. **Init** — creates the OTLP/HTTP trace, log and metric exporters, builds the shared resource from the standard env variables, installs the global tracer provider, the global logger provider, the global meter provider and the W3C trace-context propagator, then wraps `http.DefaultTransport` for outgoing request tracing.
2. **Close** — shuts down all three providers so buffered spans, log records and metrics are exported. Uses a fresh 5 second context because the lifecycle context is already canceled at this point.
