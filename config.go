package app

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

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

// validateAppName checks if name has upper camelcase format.
func validateAppName(n string) bool {
	regex := regexp.MustCompile("^(?:[A-Z][A-Za-z]*)+$")

	return regex.MatchString(n)
}

func splitUpperCamelCase(s string) []string {
	if s == "" {
		return make([]string, 0)
	}

	if len(s) == 1 {
		return []string{s}
	}

	runes := []rune(s)
	var words []string
	start := 0

	for i := 1; i < len(runes); i++ {
		if unicode.IsLower(runes[i-1]) && unicode.IsUpper(runes[i]) {
			words = append(words, s[start:i])
			start = i
		}
	}

	words = append(words, s[start:])

	return words
}
