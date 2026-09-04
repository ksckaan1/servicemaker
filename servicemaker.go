package servicemaker

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/caarlos0/env/v11"
	"github.com/ksckaan1/logger"
	"golang.org/x/sync/errgroup"
)

type ServiceMaker struct {
	comps    map[string]any
	gCtx     context.Context
	closers  []func(context.Context) error
	closerWg sync.WaitGroup
}

func New(ctx context.Context) *ServiceMaker {
	sm := &ServiceMaker{
		comps:    map[string]any{},
		closerWg: sync.WaitGroup{},
	}

	sm.gCtx = sm.createGracefulShutdown(ctx)

	return sm
}

func (s *ServiceMaker) Register[T any]() T {
	var c T

	err := s.parseConfig(&c)
	if err != nil {
		logger.Default.Fatal(
			s.gCtx, "error when parsing config",
			"component", fmt.Sprintf("%T", c),
			"error", err,
		)
		return c
	}

	compInitializer, ok := any(&c).(Initializer)
	if ok {
		err := compInitializer.Init(s.gCtx)
		if err != nil {
			logger.Default.Fatal(
				s.gCtx, "error when initializing component",
				"error", fmt.Errorf("error when initializing component: %T", compInitializer),
			)
			return c
		}
	}

	name := fmt.Sprintf("%T", &c)
	s.comps[name] = &c

	if closer, ok := any(&c).(Closer); ok {
		s.closers = append(s.closers, closer.Close)
		s.closerWg.Add(1)
	}

	return c
}

func (s *ServiceMaker) Get[T any]() *T {
	var c *T
	name := fmt.Sprintf("%T", c)
	comp, ok := s.comps[name]
	if !ok {
		logger.Default.Fatal(
			s.gCtx, "error when getting component",
			"error", fmt.Errorf("component not found: %T", c),
		)
		return nil
	}
	return comp.(*T)
}

func (s *ServiceMaker) Run() error {
	eg, ctx := errgroup.WithContext(s.gCtx)

	for _, comp := range s.comps {
		compRunner, ok := comp.(Runner)
		if !ok {
			continue
		}

		eg.Go(func() error {
			err := compRunner.Run(ctx)
			if err != nil {
				return fmt.Errorf("error when running component: %w", err)
			}
			return nil
		})
	}

	err := eg.Wait()
	if err != nil {
		return err
	}

	s.closerWg.Wait()

	return nil
}

func (s *ServiceMaker) createGracefulShutdown(ctx context.Context) context.Context {
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

func (s *ServiceMaker) parseConfig(c any) error {
	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("env.Parse: %w (%T)", err, c)
	}

	return nil
}
