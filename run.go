package servicemaker

import (
	"fmt"

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
			err := compRunner.Run(ctx)
			if err != nil {
				return fmt.Errorf("error when running component: %w", err)
			}
			return nil
		})

		return true
	})

	err := eg.Wait()
	if err != nil {
		s.closeAll()
		return err
	}

	s.closerWg.Wait()

	return nil
}
