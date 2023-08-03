package logger

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"strings"
	"time"
)

// duplicate strings in code
const (
	contentType = "Content-Type"
	userAgent   = "User-Agent"
	accept      = "Accept"
)

// exceptions - User-Agent exceptions
var (
	exception  bool
	exceptions = [...]string{
		"kube-probe", "Prometheus",
	}
)

// Logger -
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		t1 := time.Now()
		defer func() {
			for i := range exceptions {
				if strings.HasPrefix(c.Request.Header.Get(userAgent), exceptions[i]) {
					exception = true
				}
			}
			if !exception {
				requestParams := zerolog.Dict()
				for k, v := range c.Request.URL.Query() {
					requestParams.Strs(k, v)
				}
				log.Info().
					Int("return-code", c.Writer.Status()).
					Str("X-Trace-Id", c.Writer.Header().Get("X-Trace-Id")).
					Dict("message", zerolog.Dict().
						Str("http-method", c.Request.Method).
						Dict("request-headers", zerolog.Dict().
							Str(contentType, c.Request.Header.Get(contentType)).
							Str(userAgent, c.Request.Header.Get(userAgent)).
							Str(accept, c.Request.Header.Get(accept)),
						).
						Str("Path", c.Request.URL.Path).
						Dur("Latency", time.Since(t1)).
						Dict("request-params", requestParams),
					).Msg("")
			}
			exception = false
		}()
		c.Next()
	}
}
