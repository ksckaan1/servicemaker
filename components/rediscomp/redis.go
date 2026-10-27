package rediscomp

import (
	"context"
	"fmt"
	"sync"

	"github.com/ksckaan1/logger"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"go.opentelemetry.io/otel/attribute"
)

type Redis struct {
	mu    sync.Mutex
	conns map[int]*redis.Client

	// CONFIGS
	Addr       string `env:"REDIS_ADDR,required"`
	User       string `env:"REDIS_USER"`
	Pass       string `env:"REDIS_PASS"`
	ClientName string `env:"REDIS_CLIENT_NAME"`
}

func (r *Redis) Init(ctx context.Context) error {
	r.conns = make(map[int]*redis.Client)

	logger.Default.Info(ctx, "redis client initialized")

	return nil
}

func (r *Redis) Close(ctx context.Context) error {
	for _, conn := range r.clients() {
		err := conn.Close()
		if err != nil {
			logger.Default.Error(
				ctx, "error when closing redis client connection",
				"db", conn.Options().DB,
			)
			continue
		}
		logger.Default.Info(
			ctx, "redis client connection closed",
			"db", conn.Options().DB,
		)
	}
	return nil
}

func (r *Redis) HealthCheck(ctx context.Context) error {
	for _, conn := range r.clients() {
		_, err := conn.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("error when sending ping to redis connection (%d): %w", conn.Options().DB, err)
		}
	}
	return nil
}

// clients returns a snapshot of the open clients, so callers never hold the
// lock across network calls.
func (r *Redis) clients() []*redis.Client {
	r.mu.Lock()
	defer r.mu.Unlock()

	conns := make([]*redis.Client, 0, len(r.conns))
	for _, conn := range r.conns {
		conns = append(conns, conn)
	}
	return conns
}

func (r *Redis) DB(db int) *redis.Client {
	r.mu.Lock()
	defer r.mu.Unlock()

	if conn, ok := r.conns[db]; ok {
		return conn
	}

	conn := redis.NewClient(&redis.Options{
		Addr:       r.Addr,
		ClientName: r.ClientName,
		Username:   r.User,
		Password:   r.Pass,
		DB:         db,
	})
	r.conns[db] = conn

	err := redisotel.InstrumentTracing(
		conn,
		redisotel.WithAttributes(attribute.Int("db_num", db)),
	)
	if err != nil {
		logger.Default.Error(
			context.Background(),
			"error when instrumenting tracing for redis",
			"db", db,
			"error", err,
		)
	}

	err = redisotel.InstrumentMetrics(
		conn,
		redisotel.WithAttributes(attribute.Int("db_num", db)),
	)
	if err != nil {
		logger.Default.Error(
			context.Background(),
			"error when instrumenting metrics for redis",
			"db", db,
			"error", err,
		)
	}

	return conn
}
