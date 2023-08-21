package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	HttpRequestError = Metric{
		Name: "request_error_total",
		Collector: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "request_error_total",
			Help:      "Total number of http request execution errors",
		}, []string{"endpoint_uri", "error_code"}),
	}
	HttpRequestSuccessful = Metric{
		Name: "request_successful_total",
		Collector: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "http",
			Name:      "request_successful_total",
			Help:      "Total number of successful http request executions",
		}, []string{"endpoint_uri"}),
	}
	HttpRequestExecutionTime = Metric{
		Name: "request_execution_time_seconds",
		Collector: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "http",
			Name:      "request_execution_time_seconds",
			Help:      "Time of http request execution",
			Buckets:   []float64{0.1, 0.3, 0.5, 0.6, 1, 2, 5, 10, 20, 60},
		}, []string{"endpoint_uri"}),
	}
)

var (
	StorageRequestError = Metric{
		Name: "request_error_total",
		Collector: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "storage",
			Name:      "request_error_total",
			Help:      "Total number of storage request execution errors",
		}, []string{"type"}),
	}
	StorageRequestSuccessful = Metric{
		Name: "request_successful_total",
		Collector: prometheus.NewCounterVec(prometheus.CounterOpts{
			Subsystem: "storage",
			Name:      "request_successful_total",
			Help:      "Total number of successful storage request executions",
		}, []string{"type"}),
	}
	StorageRequestExecutionTime = Metric{
		Name: "request_execution_time_seconds",
		Collector: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Subsystem: "storage",
			Name:      "request_execution_time_seconds",
			Help:      "Time of storage request execution",
			Buckets:   []float64{0.1, 0.3, 0.5, 0.6, 1, 2, 5, 10, 20, 60},
		}, []string{"type"}),
	}
)
