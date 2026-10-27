package servicemaker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (s *ServiceMaker) gracefulShutdown(ctx context.Context) context.Context {
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	// Release the signal registration once the context is done.
	context.AfterFunc(gCtx, cancel)

	return gCtx
}

func (s *ServiceMaker) RegisterGracefulShutdown(closer func(context.Context) error) {
	s.closers = append(s.closers, closer)
	s.closerWg.Add(1)
}
