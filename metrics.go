package app

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

func NewMetrics(prefix string) *Metrics {
	return &Metrics{
		prefix:   prefix,
		registry: prometheus.NewRegistry(),
	}
}

type Metrics struct {
	prefix   string
	registry *prometheus.Registry
}

func (m Metrics) NewCounterVec(name string, help string, labelNames []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("sf_%s_%s", m.prefix, name),
		Help: help,
	}, labelNames)

	m.registry.MustRegister(c)

	return c
}
