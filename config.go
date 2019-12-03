package app

import (
	"fmt"
	"strings"

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

// BuildEnvConfigName will format an environment variable name
// string from the supplied parts.
func BuildEnvConfigName(name ...string) string {
	for i := range name {
		name[i] = strings.ToUpper(name[i])
	}
	return strings.Join(name, "_")
}
