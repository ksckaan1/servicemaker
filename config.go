package servicemaker

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

func ParseConfig(c any) error {
	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("env.Parse: %w (%T)", err, c)
	}

	return nil
}
