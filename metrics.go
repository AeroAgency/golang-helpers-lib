package helpers

import (
	"github.com/prometheus/client_golang/prometheus"
	"net/http"
	"strconv"
	"time"
)

var (
	// Create a customized counter metric.
	RequestCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "http_server_requests_seconds_count",
		Help: "Application request count.",
	}, []string{"method", "uri", "status"})
	RequestDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "http_server_requests_seconds_sum",
		Help: "Request duration, ms",
	}, []string{"method", "uri", "status"})
)

// IncRequestCount
func IncRequestCount(r *http.Request, lrw *LoggingResponseWriter) {
	RequestCount.WithLabelValues(r.Method, r.RequestURI, strconv.Itoa(lrw.StatusCode)).Inc()
}

// SetRequestDuration
func SetRequestDuration(r *http.Request, lrw *LoggingResponseWriter, startTime int) {
	durationTime := float64((time.Now().Nanosecond() - startTime) % 1000)
	RequestDuration.WithLabelValues(r.Method, r.RequestURI, strconv.Itoa(lrw.StatusCode)).Set(durationTime)
}
