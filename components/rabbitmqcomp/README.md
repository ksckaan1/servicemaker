# RabbitMQ Component

RabbitMQ client component for ServiceMaker, using [amqp091-go](https://github.com/rabbitmq/amqp091-go).

## Install

```bash
go get github.com/ksckaan1/servicemaker/components/rabbitmqcomp
```

## Usage

```go
package main

import (
    "context"
    "log"

    "github.com/ksckaan1/servicemaker"
    "github.com/ksckaan1/servicemaker/components/rabbitmqcomp"
)

func main() {
    sm := servicemaker.New(context.Background())

    rmq := sm.Register[rabbitmqcomp.RabbitMQ]()

    err := rmq.Ch().PublishWithContext(context.Background(),
        "", "my-queue", false, false,
        amqp.Publishing{ContentType: "text/plain", Body: []byte("hello")},
    )
    if err != nil {
        log.Fatal(err)
    }

    log.Fatal(sm.Run())
}
```

```bash
RMQ_CONNECTION_STRING=amqp://guest:guest@localhost:5672/ go run main.go
```

## Configuration

| Env Variable | Default | Description |
|-------------|---------|-------------|
| `RMQ_CONNECTION_STRING` | — | AMQP connection URL |
| `RMQ_PREFETCH_COUNT` | `1` | QoS prefetch count |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — dials connection, opens channel, sets QoS |
| `Runner` | No — RabbitMQ has no blocking process |
| `Closer` | Yes — closes the connection |
| `HealthChecker` | Yes — checks connection and channel state |

## Methods

### `Ch() *amqp.Channel`

Returns the underlying AMQP channel for publishing and consuming.

## Lifecycle

1. **Init** — `amqp.Dial()` connects to the server, opens a channel, and applies QoS settings.
2. **Run** — not implemented (returns immediately).
3. **Close** — closes the AMQP connection (which also closes all channels).
4. **HealthCheck** — checks if both the connection and channel are still open.
