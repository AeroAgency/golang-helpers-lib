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
func HttpMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		methodId := fmt.Sprintf("handler %s:%s", c.Request.Method, c.FullPath())
		tracer := tracerAdapter.NewTracer(methodId)
		defer tracer.Close()
		t1 := time.Now()
		var requestBody string
		if c.Request.Method == http.MethodPost {
			buf, _ := io.ReadAll(c.Request.Body)
			rdr1 := io.NopCloser(bytes.NewBuffer(buf))
			rdr2 := io.NopCloser(bytes.NewBuffer(buf)) //We have to create a new Buffer, because rdr1 will be read.
			r, _ := io.ReadAll(rdr1)
			requestBody = string(r)
			tracer.LogData("[Request body]", requestBody)
			c.Request.Body = rdr2
		}
		defer func(requestBody string) {
			for i := range exceptions {
				if strings.HasPrefix(c.Request.Header.Get(userAgent), exceptions[i]) {
					exception = true
				}
			}
			if !exception {
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
				log.Info().
					Str("log-type", "request").
					Int("return-code", c.Writer.Status()).
					Str("X-Trace-Id", c.Writer.Header().Get("X-Trace-Id")).
					Dict("message", zerolog.Dict().
						Str("http-method", c.Request.Method).
						Str("Path", c.FullPath()).
						Str(userAgent, c.Request.Header.Get(userAgent)).
						Dur("Latency", time.Since(t1)).
						Dict("request-params", requestParams).
						Str("request-body", requestBody),
					).Msg("")
			}
			exception = false
		}(requestBody)
		c.Set("tracer", tracer)
		c.Next()
	}
}

func HttpRespLogMiddleWare(logResponses bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if logResponses {
			blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
			c.Writer = blw
			c.Next()
			tracer, _ := c.MustGet("tracer").(tracerAdapter.TracerInterface)
			responseBody := blw.body.String()
			tracer.Log("[Response Body]", responseBody)
			log.Info().
				Str("log-type", "response").
				Int("return-code", c.Writer.Status()).
				Str("X-Trace-Id", c.Writer.Header().Get("X-Trace-Id")).
				Dict("message", zerolog.Dict().
					Str("response-body", responseBody),
				).Msg("")

		} else {
			c.Next()
		}
	}
}
