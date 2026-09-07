package servicemaker

import (
	"context"

	"github.com/ksckaan1/logger"
)

func (s *ServiceMaker) closeAll() {
	ctx := context.Background()
	for _, closer := range s.closers {
		err := closer(ctx)
		if err != nil {
			logger.Default.Error(ctx, "error when closing component", "error", err)
		}
		s.closerWg.Done()
	}
}
