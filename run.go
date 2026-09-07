package servicemaker

import (
	"context"
	"fmt"

	"github.com/ksckaan1/logger"
	"golang.org/x/sync/errgroup"
)

func (s *ServiceMaker) Run() error {
	eg, ctx := errgroup.WithContext(s.gCtx)

	s.comps.Range(func(_, comp any) bool {
		compRunner, ok := comp.(Runner)
		if !ok {
			return true
		}

		eg.Go(func() error {
			done := make(chan error, 1)
			go func() {
				done <- compRunner.Run(ctx)
			}()
			select {
			case err := <-done:
				return fmt.Errorf("(%T).Run: %w", comp, err)
			case <-ctx.Done():
				return ctx.Err()
			}
		})

		return true
	})

	err := eg.Wait()

	s.closeAll()
	s.closerWg.Wait()

	select {
	case <-s.shutdownCh:
		logger.Default.Info(context.Background(), "graceful shutdown completed")
		return nil
	default:
		return err
	}
}
