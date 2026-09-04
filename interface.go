package servicemaker

import "context"

type Initializer interface {
	Init(ctx context.Context) error
}

type Runner interface {
	Run(ctx context.Context) error
}

type Closer interface {
	Close(ctx context.Context) error
}

type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}
