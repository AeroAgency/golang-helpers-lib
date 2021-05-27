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
	contentType       = "Content-Type"
	contentLength     = "Content-Length"
	userAgent         = "User-Agent"
	server            = "Server"
	via               = "Via"
	accept            = "Accept"
	pstxid            = "pstxid"
	xPsPstxid         = "x-ps-pstxid"
	xPsSsoToken       = "x-ps-sso-token"
	frontEndRequestID = "FrontEnd-Request-ID"
	xForwardedFor     = "X-FORWARDED-FOR"
	employeeNumber    = "employee-number"
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

				log.Info().Dict("message", zerolog.Dict().
					Int("return-code", c.Writer.Status()).
					Str("http-method", c.Request.Method).
					Dict("request-headers", zerolog.Dict().
						Str(contentType, c.Request.Header.Get(contentType)).
						Str(contentLength, c.Request.Header.Get(contentLength)).
						Str(userAgent, c.Request.Header.Get(userAgent)).
						Str(server, c.Request.Header.Get(server)).
						Str(via, c.Request.Header.Get(via)).
						Str(accept, c.Request.Header.Get(accept)).
						Str(pstxid, c.Request.Header.Get(pstxid)).
						Str(xPsPstxid, c.Request.Header.Get(xPsPstxid)).
						Str(xPsSsoToken, c.Request.Header.Get(xPsSsoToken)).
						Str(frontEndRequestID, c.Request.Header.Get(frontEndRequestID)),
					).
					Str(xForwardedFor, c.Request.Header.Get(xForwardedFor)).
					Str("Remote Addr", c.Request.RemoteAddr).
					Str("Proto", c.Request.Proto).
					Str("Path", c.Request.URL.Path).
					Dur("Latency", time.Since(t1)).
					Str(employeeNumber, c.Request.Header.Get(employeeNumber)).
					Dict("request-params", requestParams),
				).Msg("")
			}

			exception = false
		}()

		c.Next()
	}
}
