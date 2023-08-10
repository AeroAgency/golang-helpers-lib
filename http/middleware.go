package http

import (
	"bytes"
	"fmt"
	tracerAdapter "github.com/AeroAgency/go-gin-tracer"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"strings"
	"time"
)

// duplicate strings in code
const (
	userAgent = "User-Agent"
)

// exceptions - User-Agent exceptions
var (
	exception  bool
	exceptions = [...]string{
		"kube-probe", "Prometheus",
	}
)

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// HttpMiddleWare -
func HttpMiddleWare(logResponses bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		methodId := fmt.Sprintf("handler %s:%s", c.Request.Method, c.FullPath())
		tracer := tracerAdapter.NewTracer(methodId)
		defer tracer.Close()
		t1 := time.Now()
		defer func() {
			for i := range exceptions {
				if strings.HasPrefix(c.Request.Header.Get(userAgent), exceptions[i]) {
					exception = true
				}
			}
			if !exception {
				if c.Request.Method == http.MethodPost {
					buf, _ := io.ReadAll(c.Request.Body)
					rdr1 := io.NopCloser(bytes.NewBuffer(buf))
					rdr2 := io.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.
					tracer.LogData("[Request body]", rdr1)
					c.Request.Body = rdr2
				}
				requestParams := zerolog.Dict()
				allQueryParams := c.Request.URL.Query()
				for k, v := range allQueryParams {
					requestParams.Strs(k, v)
					queryVal := strings.Join(v, ", ")
					tracer.Log(fmt.Sprintf("Query-param [%s]", k), queryVal)
				}
				for k, values := range c.Request.Header {
					// Loop over all values for the name.
					headerVal := strings.Join(values, ", ")
					tracer.Log(fmt.Sprintf("Header [%s]", k), headerVal)
				}

				var responseBody string
				if logResponses {
					blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
					c.Writer = blw
					responseBody = blw.body.String()
				}

				log.Info().
					Int("return-code", c.Writer.Status()).
					Str("X-Trace-Id", c.Writer.Header().Get("X-Trace-Id")).
					Dict("message", zerolog.Dict().
						Str("http-method", c.Request.Method).
						Str("Path", c.FullPath()).
						Dur("Latency", time.Since(t1)).
						Dict("request-params", requestParams).
						Str("response-body", responseBody),
					).Msg("")
			}
			exception = false
		}()
		c.Set("tracer", tracer)
		c.Next()
	}
}
