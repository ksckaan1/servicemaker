package servicemaker

import (
	"fmt"

	"github.com/caarlos0/env/v11"
)

func ParseConfig(c any, prefix ...string) error {
	if len(prefix) > 0 {
		err := env.ParseWithOptions(c, env.Options{
			Prefix: prefix[0],
		})
		if err != nil {
			return fmt.Errorf("env.ParseWithOptions: %w (%T)", err, c)
		}
		return nil
	}

	err := env.Parse(c)
	if err != nil {
		return fmt.Errorf("env.Parse: %w (%T)", err, c)
	}

	return nil
}
