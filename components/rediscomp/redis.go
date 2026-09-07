package rediscomp

import (
	"context"
	"fmt"

	"github.com/ksckaan1/logger"
	"github.com/redis/go-redis/v9"
)

type Redis struct {
	conns map[int]*redis.Client

	// CONFIGS
	Addr       string `env:"REDIS_ADDR"`
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
	for _, conn := range r.conns {
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
	for _, conn := range r.conns {
		_, err := conn.Ping(ctx).Result()
		if err != nil {
			return fmt.Errorf("error when sending ping to redis connection (%d): %w", conn.Options().DB, err)
		}
	}
	return nil
}

func (r *Redis) DB(db int) *redis.Client {
	conn, ok := r.conns[db]
	if !ok {
		conn = redis.NewClient(&redis.Options{
			Addr:       r.Addr,
			ClientName: r.ClientName,
			Username:   r.User,
			Password:   r.Pass,
			DB:         db,
		})
		r.conns[db] = conn
	}

	return conn
}
