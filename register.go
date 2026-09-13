package servicemaker

import (
	"fmt"
	"strings"

	"github.com/ksckaan1/logger"
)

func (s *ServiceMaker) Register[T any]() *T {
	return s.RegisterNamed[T]("")
}

func (s *ServiceMaker) RegisterNamed[T any](name string) *T {
	var c T

	var envPrefix []string
	var prefix string

	if name != "" {
		envPrefix = []string{fmt.Sprintf("%s_", strings.ToUpper(name))}
		prefix = fmt.Sprintf("%s_", name)
	}

	key := fmt.Sprintf("%s%T", prefix, &c)

	err := ParseConfig(&c, envPrefix...)
	if err != nil {
		logger.Default.Fatal(
			s.gCtx, "error when parsing config",
			"component", key,
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
				"error", fmt.Errorf("error when initializing component (%s): %w", key, err),
			)
			return &c
		}
	}

	_, ok = s.comps.Load(key)
	if ok {
		logger.Default.Fatal(
			s.gCtx, "error when registering component",
			"component", key,
			"error", fmt.Errorf("component %s already registered", key),
		)
	}

	s.comps.Store(key, &c)

	if closer, ok := any(&c).(Closer); ok {
		s.closers = append(s.closers, closer.Close)
		s.closerWg.Add(1)
	}

	return &c
}
