package app

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadEnvConfig(t *testing.T) {
	c := &struct {
		Foo string
	}{}

	os.Setenv("MYPREFIX_FOO", "Foo")
	defer os.Unsetenv("MYPREFIX_FOO")

	err := ReadEnvConfig("MYPREFIX", c)
	assert.Nil(t, err)
	assert.Equal(t, "Foo", c.Foo)
}

func TestBuildEnvConfigName(t *testing.T) {
	assert.Equal(t, "MYAPP_FOO_BAR", BuildEnvConfigName("MyApp", "foo", "BAR"))
}
