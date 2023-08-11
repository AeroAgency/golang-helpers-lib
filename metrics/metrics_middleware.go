package metrics

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

func Middleware(m Metrics) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		defer func() {
			_ = m.Observe("request_execution_time_seconds", time.Since(start).Seconds(), c.Request.URL.Path)
		}()

		c.Next()
		statusCode := c.Writer.Status()

		if statusCode != http.StatusOK {
			_ = m.Inc("request_error_total", c.Request.URL.Path, strconv.Itoa(statusCode))
			return
		}
		_ = m.Inc("request_successful_total", c.Request.URL.Path)
	}
}
