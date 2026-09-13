package servicemaker

import (
	"fmt"

	"github.com/ksckaan1/logger"
)

func (s *ServiceMaker) Get[T any]() *T {
	var c *T
	name := fmt.Sprintf("%T", c)
	comp, ok := s.comps.Load(name)
	if !ok {
		logger.Default.Fatal(
			s.gCtx, "error when getting component",
			"error", fmt.Errorf("component not found: %T", c),
		)
		return nil
	}
	return comp.(*T)
}

func (s *ServiceMaker) GetNamed[T any](name string) *T {
	var c *T
	key := fmt.Sprintf("%s_%T", name, c)
	comp, ok := s.comps.Load(key)
	if !ok {
		logger.Default.Fatal(
			s.gCtx, "error when getting component",
			"error", fmt.Errorf("component not found: %s", key),
		)
		return nil
	}
	return comp.(*T)
}
