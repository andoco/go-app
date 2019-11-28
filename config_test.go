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

	err := ReadEnvConfig("MYPREFIX", c)
	assert.Nil(t, err)
	assert.Equal(t, "Foo", c.Foo)
}
