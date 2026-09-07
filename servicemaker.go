package servicemaker

import (
	"context"
	"sync"
)

type ServiceMaker struct {
	comps    sync.Map // component container
	gCtx     context.Context
	closers  []func(context.Context) error
	closerWg sync.WaitGroup
	closerMu sync.Mutex
	closed   bool
}

func New(ctx context.Context) *ServiceMaker {
	sm := &ServiceMaker{
		comps:    sync.Map{},
		closerWg: sync.WaitGroup{},
	}

	sm.gCtx = sm.gracefulShutdown(ctx)

	return sm
}
