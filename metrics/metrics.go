package metrics

import (
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
)

var PullMode bool

type Metric struct {
	Name      string
	Collector prometheus.Collector
}

type Metrics interface {
	Inc(metricName string, labelValues ...string) error
	Observe(metricName string, value float64, labelValues ...string) error
}

type PullMetrics struct {
	metrics map[string]prometheus.Collector
}

func NewPullMetrics(reg prometheus.Registerer, metrics []Metric) *PullMetrics {
	m := &PullMetrics{
		metrics: make(map[string]prometheus.Collector),
	}

	for _, metric := range metrics {
		m.metrics[metric.Name] = metric.Collector
	}

	for _, metric := range m.metrics {
		reg.MustRegister(metric)
	}

	return m
}

func (m *PullMetrics) Inc(metricName string, labelValues ...string) error {
	metric, ok := m.metrics[metricName]

	if !ok {
		return errors.Errorf("metric '%s' not existed.", metricName)
	}

	if err := inc(metric, labelValues); err != nil {
		return err
	}

	return nil
}

func (m *PullMetrics) Observe(metricName string, value float64, labelValues ...string) error {
	metric, ok := m.metrics[metricName]

	if !ok {
		return errors.Errorf("metric '%s' not existed.", metricName)
	}

	if err := observe(metric, value, labelValues); err != nil {
		return err
	}

	return nil
}

func inc(metric prometheus.Collector, labelValues []string) error {
	switch metric := metric.(type) {
	case *prometheus.CounterVec:
		metric.WithLabelValues(labelValues...).Inc()
	case prometheus.Counter:
		metric.Inc()
	case *prometheus.GaugeVec:
		metric.WithLabelValues(labelValues...).Inc()
	case prometheus.Gauge:
		metric.Inc()
	default:
		return errors.Errorf("metric is not Gauge or Counter type")
	}

	return nil
}

func observe(metric prometheus.Collector, value float64, labelValues []string) error {
	switch metric := metric.(type) {
	case *prometheus.HistogramVec:
		metric.WithLabelValues(labelValues...).Observe(value)
	case prometheus.Histogram:
		metric.Observe(value)
	case *prometheus.SummaryVec:
		metric.WithLabelValues(labelValues...).Observe(value)
	case prometheus.Summary:
		metric.Observe(value)
	default:
		return errors.Errorf("metric is not Histogram or Summary type")
	}
	return nil
}
