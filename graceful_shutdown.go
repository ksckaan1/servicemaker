package servicemaker

import (
	"context"
	"os"
	"os/signal"

	"github.com/ksckaan1/logger"
)

func (s *ServiceMaker) gracefulShutdown(ctx context.Context) context.Context {
	gCtx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill)
	go func() {
		<-gCtx.Done()
		defer cancel()
		for _, closer := range s.closers {
			err := closer(gCtx)
			if err != nil {
				logger.Default.Error(gCtx, "error when closing component", "error", err)
			}
			s.closerWg.Done()
		}
	}()
	return gCtx
}
