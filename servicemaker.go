package servicemaker

import (
	"context"
	"sync"
)

type ServiceMaker struct {
	comps      sync.Map
	gCtx       context.Context
	closers    []func(context.Context) error
	closerWg   sync.WaitGroup
	shutdownCh chan struct{}
}

func New(ctx context.Context) *ServiceMaker {
	sm := &ServiceMaker{
		shutdownCh: make(chan struct{}),
	}

	sm.gCtx = sm.gracefulShutdown(ctx)

	return sm
}
