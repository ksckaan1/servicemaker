# MeiliSearch Component `v0.3.0`

MeiliSearch client component for ServiceMaker, using [meilisearch-go](https://github.com/meilisearch/meilisearch-go).

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/meilisearchcomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/meilisearchcomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    ms := sm.Register[meilisearchcomp.MeiliSearch]()

    index := ms.ServiceManager().Index("movies")
    _, err := index.AddDocuments(ctx, []map[string]any{
        {"id": 1, "title": "Toy Story"},
        {"id": 2, "title": "Cars"},
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(sm.Run())
}
```

```bash
MS_HOST=http://localhost:7700 MS_API_KEY=masterKey go run main.go
```

## Configuration

All configuration is read from environment variables.

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `MS_HOST` | **required** | MeiliSearch server address (e.g. `http://localhost:7700`) |
| `MS_API_KEY` | — | API key (optional, omitted if empty) |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — creates the MeiliSearch client |
| `Runner` | No — MeiliSearch has no blocking process |
| `Closer` | Yes — closes the client connection |
| `HealthChecker` | Yes — checks server health |

## Methods

### `ServiceManager() meilisearch.ServiceManager`

Returns the underlying MeiliSearch service manager for index operations.

```go
ms := sm.Register[meilisearchcomp.MeiliSearch]()

index := ms.ServiceManager().Index("my-index")
```

## Lifecycle

1. **Init** — `meilisearch.New()` creates the client with the provided host and optional API key.
2. **Run** — not implemented (returns immediately).
3. **Close** — `ServiceManager.Close()` shuts down the client connection.
4. **HealthCheck** — `HealthWithContext()` verifies the server is reachable.

## OpenTelemetry Support

Instrumentation is active only when `otelcomp.Otel` is registered — without it every provider is no-op. The Meilisearch client is given a custom `http.Client` whose transport is wrapped with `otelhttp`, so every API call is traced and measured.

### Trace

| Span | Kind | What it covers |
|------|------|----------------|
| `GET`, `POST`, `DELETE`, … | `client` | One per HTTP request to Meilisearch |

| Attribute | Example |
|-----------|---------|
| `http.request.method` | `GET` |
| `url.full` | `http://localhost:7700/indexes` |
| `url.scheme` | `http` |
| `http.response.status_code` | `200` |
| `server.address`, `server.port` | `localhost`, `7700` |
| `network.protocol.version` | `1.1` |
| `error.type` | on a transport failure |

Retries performed by the Meilisearch client produce one span each.

### Metric

| Metric | Type | Unit |
|--------|------|------|
| `http.client.request.duration` | histogram | `s` |
| `http.client.request.body.size` | histogram | `By` |
| `http.client.response.body.size` | histogram | `By` |

Attributes: `http.request.method`, `http.response.status_code`, `server.address`, `server.port`, `url.scheme`.
