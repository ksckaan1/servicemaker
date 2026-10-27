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

    // Each caller needs its own channel: amqp.Channel is not safe for
    // concurrent use.
    ch, err := rmq.NewChannel()
    if err != nil {
        log.Fatal(err)
    }
    defer ch.Close()

    err = ch.PublishWithContext(context.Background(),
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
| `RMQ_CONNECTION_STRING` | **required** | AMQP connection URL |
| `RMQ_PREFETCH_COUNT` | `1` | QoS prefetch count |

## Interfaces

| Interface | Implemented |
|-----------|-------------|
| `Initializer` | Yes — dials the connection |
| `Runner` | No — RabbitMQ has no blocking process |
| `Closer` | Yes — closes the connection |
| `HealthChecker` | Yes — checks the connection state |

## Methods

### `NewChannel() (*amqp.Channel, error)`

Opens a channel on the current connection and applies `RMQ_PREFETCH_COUNT` as
the QoS prefetch. Call it wherever you publish or consume, and close the
channel yourself.

Every caller needs its own channel — a single `amqp.Channel` is not safe for
concurrent use, so sharing one between handlers corrupts the frame stream and
closes the channel.

If the broker already closed the connection (a restart, or an idle timeout),
`NewChannel` dials a new one first, so publishes recover without restarting
the process. Earlier channels on the dead connection do not recover; open a
new one.

## Lifecycle

1. **Init** — `amqp.Dial()` connects to the server.
2. **Run** — not implemented (returns immediately).
3. **Close** — closes the AMQP connection (which also closes all channels).
4. **HealthCheck** — checks if the connection is still open.
