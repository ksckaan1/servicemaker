package servicemaker

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func (s *ServiceMaker) gracefulShutdown(ctx context.Context) context.Context {
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-gCtx.Done()
		defer cancel()
		close(s.shutdownCh)
	}()
	return gCtx
}
