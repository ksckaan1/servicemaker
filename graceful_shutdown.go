package servicemaker

import (
	"context"
	"os"
	"os/signal"
)

func (s *ServiceMaker) gracefulShutdown(ctx context.Context) context.Context {
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	go func() {
		<-gCtx.Done()
		defer cancel()
		s.closeAll()
	}()
	return gCtx
}
