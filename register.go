package servicemaker

import (
	"fmt"

	"github.com/ksckaan1/logger"
)

func (s *ServiceMaker) Register[T any]() *T {
	var c T

	err := ParseConfig(&c)
	if err != nil {
		logger.Default.Fatal(
			s.gCtx, "error when parsing config",
			"component", fmt.Sprintf("%T", c),
			"error", err,
		)
		return &c
	}

	compInitializer, ok := any(&c).(Initializer)
	if ok {
		err := compInitializer.Init(s.gCtx)
		if err != nil {
			logger.Default.Fatal(
				s.gCtx, "error when initializing component",
				"error", fmt.Errorf("error when initializing component (%T): %w", compInitializer, err),
			)
			return &c
		}
	}

	name := fmt.Sprintf("%T", &c)

	_, ok = s.comps.Load(name)
	if ok {
		logger.Default.Fatal(
			s.gCtx, "error when registering component",
			"component", name,
			"error", fmt.Errorf("component %s already registered", name),
		)
	}

	s.comps.Store(name, &c)

	if closer, ok := any(&c).(Closer); ok {
		s.closers = append(s.closers, closer.Close)
		s.closerWg.Add(1)
	}

	return &c
}
