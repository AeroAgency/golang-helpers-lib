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
	CalculationsDocumentDownloads = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dealer_portal_download_prm_reports",
		Help: "DealerPortal.DownloadPrmReports",
	}, []string{"UserDownloadResult"}) // SUCCESS, FAILED
	AuthSignIn = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dealer_portal_authorizations",
		Help: "DealerPortal.authrizations",
	}, []string{"result"}) // 200 или 401
	RequestsRequests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "dealer_portal_requests_status",
		Help: "DealerPortal.RequestsStatus",
	}, []string{"UserDownloadResult"}) // DP_to_HPSM_OK, DP_to_HPSM_ERROR, HPSM_to_DP_OK, HPSM_to_DP_ERROR
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

// IncCalculationsReportDownloads
func IncCalculationsReportDownloads(result string) {
	CalculationsDocumentDownloads.WithLabelValues(result).Inc()
}

// IncAuthSignIn
func IncAuthSignIn(result string) {
	AuthSignIn.WithLabelValues(result).Inc()
}

// IncRequestsRequests
func IncRequestsRequests(result string) {
	RequestsRequests.WithLabelValues(result).Inc()
}
