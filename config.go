package app

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// ReadEnvConfig will read config from environment variables
// prefixed with prefix and set values on c.
func ReadEnvConfig(prefix string, c interface{}) error {
	if err := envconfig.Process(prefix, c); err != nil {
		return fmt.Errorf("could not read config: %w", err)
	}
	return nil
}
