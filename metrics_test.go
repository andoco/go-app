package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics("foo_bar")

	assert.Equal(t, "foo_bar", m.prefix)
}
