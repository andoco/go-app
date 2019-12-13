package app

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

type PrometheusConfig struct {
	Enabled bool
	Path    string `default:"/metrics"`
	Port    int    `default:"9090"`
	Prefix  string
}

func NewMetrics(config PrometheusConfig) *Metrics {
	return &Metrics{
		prefix:   config.Prefix,
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
