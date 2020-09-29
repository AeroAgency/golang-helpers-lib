package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

type PrometheusMetrics struct {
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.GaugeVec
}

func NewPrometheusMetrics() *PrometheusMetrics {
	return &PrometheusMetrics{
		requestCount: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_server_requests_seconds_count",
			Help: "Application request count.",
		}, []string{"method", "uri", "status"}),
		requestDuration: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name: "http_server_requests_seconds_sum",
			Help: "Request duration, ms",
		}, []string{"method", "uri", "status"}),
	}
}

// IncRequestCount
func (p *PrometheusMetrics) IncRequestCount(r *http.Request, lrw *LoggingResponseWriter) {
	p.requestCount.WithLabelValues(r.Method, r.RequestURI, strconv.Itoa(lrw.StatusCode)).Inc()
}

// SetRequestDuration
func (p *PrometheusMetrics) SetRequestDuration(r *http.Request, lrw *LoggingResponseWriter, startTime int) {
	durationTime := float64((time.Now().Nanosecond() - startTime) % 1000)
	p.requestDuration.WithLabelValues(r.Method, r.RequestURI, strconv.Itoa(lrw.StatusCode)).Set(durationTime)
}
