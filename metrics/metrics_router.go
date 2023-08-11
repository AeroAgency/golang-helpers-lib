package metrics

import (
	tracer "github.com/AeroAgency/go-gin-tracer"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsRouter struct {
	reg prometheus.Gatherer
}

// NewMetricsRouter Конструктор
func NewMetricsRouter(registry prometheus.Gatherer) *MetricsRouter {
	return &MetricsRouter{
		reg: registry,
	}
}

func (p MetricsRouter) Router() *gin.Engine {
	r := gin.New()
	r.Use(tracer.OpenTracingMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.HandlerFor(p.reg, promhttp.HandlerOpts{})))

	return r
}
