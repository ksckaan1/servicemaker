package servicemaker

import (
	"context"
	"sync"
)

type ServiceMaker struct {
	comps    sync.Map
	gCtx     context.Context
	closers  []func(context.Context) error
	closerWg sync.WaitGroup
}

func New(ctx context.Context) *ServiceMaker {
	sm := &ServiceMaker{}

	sm.gCtx = sm.gracefulShutdown(ctx)

	return sm
}
