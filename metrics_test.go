package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMetrics(t *testing.T) {
	m := NewMetrics(PrometheusConfig{Prefix: "foo_bar"})

	assert.Equal(t, "foo_bar", m.prefix)
}

func TestMetricName(t *testing.T) {
	assert.Equal(t, "bar", metricName("", "bar"))
	assert.Equal(t, "foo_bar", metricName("foo", "bar"))
}
