package rabbitmqcomp

import (
	"context"
	"fmt"
	"sync"

	"github.com/ksckaan1/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	mu   sync.Mutex
	conn *amqp.Connection

	// CONFIGS
	ConnectionString string `env:"RMQ_CONNECTION_STRING,required"`
	PrefetchCount    int    `env:"RMQ_PREFETCH_COUNT" envDefault:"1"`
}

func (r *RabbitMQ) Init(ctx context.Context) error {
	if _, err := r.connection(); err != nil {
		return err
	}

	logger.Default.Info(ctx, "rabbitmq client initialized")

	return nil
}

func (r *RabbitMQ) Close(ctx context.Context) error {
	r.mu.Lock()
	conn := r.conn
	r.conn = nil
	r.mu.Unlock()

	if conn == nil {
		return nil
	}

	err := conn.Close()
	if err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}

	logger.Default.Info(ctx, "rabbitmq client connection closed")

	return nil
}

func (r *RabbitMQ) HealthCheck(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn == nil || r.conn.IsClosed() {
		return fmt.Errorf("amqp connection is closed")
	}

	return nil
}

// NewChannel opens a channel on the current connection, reconnecting first if
// the broker dropped it. Every caller needs its own channel, because
// amqp.Channel is not safe for concurrent use.
func (r *RabbitMQ) NewChannel() (*amqp.Channel, error) {
	conn, err := r.connection()
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("conn.Channel: %w", err)
	}

	err = ch.Qos(
		r.PrefetchCount,
		0,
		false,
	)
	if err != nil {
		return nil, fmt.Errorf("ch.Qos: %w", err)
	}

	return ch, nil
}

// connection returns a live connection, dialing a new one when there is none
// or the broker already closed it.
func (r *RabbitMQ) connection() (*amqp.Connection, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.conn != nil && !r.conn.IsClosed() {
		return r.conn, nil
	}

	conn, err := amqp.Dial(r.ConnectionString)
	if err != nil {
		return nil, fmt.Errorf("amqp.Dial: %w", err)
	}
	r.conn = conn

	return conn, nil
}
