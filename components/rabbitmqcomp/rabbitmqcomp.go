package rabbitmqcomp

import (
	"context"
	"fmt"

	"github.com/ksckaan1/logger"
	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQ struct {
	conn *amqp.Connection
	ch   *amqp.Channel

	// CONFIGS
	ConnectionString string `env:"RMQ_CONNECTION_STRING"`
	PrefetchCount    int    `env:"RMQ_PREFETCH_COUNT" envDefault:"1"`
}

func (r *RabbitMQ) Init(ctx context.Context) error {
	conn, err := amqp.Dial(r.ConnectionString)
	if err != nil {
		return fmt.Errorf("amqp.Dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("amqp.Channel: %w", err)
	}

	err = ch.Qos(
		r.PrefetchCount,
		0,
		false,
	)
	if err != nil {
		return fmt.Errorf("ch.Qos: %w", err)
	}

	r.conn = conn
	r.ch = ch

	logger.Default.Info(ctx, "rabbitmq client initialized")

	return nil
}

func (r *RabbitMQ) Close(ctx context.Context) error {
	err := r.conn.Close()
	if err != nil {
		return fmt.Errorf("conn.Close: %w", err)
	}

	logger.Default.Info(ctx, "rabbitmq client connection closed")

	return nil
}

func (r *RabbitMQ) HealthCheck(ctx context.Context) error {
	if r.conn.IsClosed() {
		return fmt.Errorf("amqp connection is closed")
	}

	if r.ch.IsClosed() {
		return fmt.Errorf("amqp channel is closed")
	}

	return nil
}

func (r *RabbitMQ) Ch() *amqp.Channel {
	return r.ch
}
