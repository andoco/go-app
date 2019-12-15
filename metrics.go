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
	r := prometheus.NewRegistry()
	r.MustRegister(prometheus.NewBuildInfoCollector())
	r.MustRegister(prometheus.NewGoCollector())

	return &Metrics{
		prefix:   config.Prefix,
		registry: r,
	}
}

type Metrics struct {
	prefix   string
	registry *prometheus.Registry
}

func (m Metrics) NewCounterVec(name string, help string, labelNames []string) *prometheus.CounterVec {
	c := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_%s", m.prefix, name),
		Help: help,
	}, labelNames)

	m.registry.MustRegister(c)

	return c
}

func (m Metrics) NewHistogramVec(name string, help string, labelNames []string) *prometheus.HistogramVec {
	c := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: fmt.Sprintf("%s_%s", m.prefix, name),
		Help: help,
	}, labelNames)

	m.registry.MustRegister(c)

	return c
}

func metricName(prefix, name string) string {
	if prefix == "" {
		return name
	}

	return fmt.Sprintf("%s_%s", prefix, name)
}
