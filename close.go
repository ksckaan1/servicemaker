package servicemaker

import (
	"context"
	"time"

	"github.com/ksckaan1/logger"
)

// closeTimeout bounds the whole close sequence, so a component that hangs
// while closing cannot keep the process alive until the orchestrator kills it.
var closeTimeout = 10 * time.Second

func (s *ServiceMaker) closeAll() {
	ctx, cancel := context.WithTimeout(context.Background(), closeTimeout)
	defer cancel()

	for _, closer := range s.closers {
		err := closer(ctx)
		if err != nil {
			logger.Default.Error(ctx, "error when closing component", "error", err)
		}
		s.closerWg.Done()
	}
}
