package redis

import (
	"context"
	"fmt"
)

type Redis struct{}

func (r *Redis) Init(ctx context.Context) error {
	fmt.Println("redis initialized")
	return nil
}

func (r *Redis) Run(ctx context.Context) error {
	fmt.Println("redis running")
	return nil
}
