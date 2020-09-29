package helpers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/protobuf/runtime/protoiface"
	"net/http"
	"strings"
	"time"
)

// userAgent - duplicate string in code
const userAgent = "User-Agent"

type Middleware struct{}

// Установка заголовков для CORS-запросов
func (m *Middleware) SetCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With,content-type,Access-Token")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

// Обработчик для ответа
func (m *Middleware) HttpResponseModifier(ctx context.Context, w http.ResponseWriter, proto protoiface.MessageV1) error {
	// очистка ненужных заголовков
	delete(w.Header(), "Grpc-Metadata-Content-Type")
	return nil
}

// Мэппинг для заголовков
func (m *Middleware) CustomMatcher(key string) (string, bool) {
	switch key {
	case "X-Request-Id":
		return key, true
	case "Privileges":
		return key, true
	case "Access-Token":
		return key, true
	default:
		return runtime.DefaultHeaderMatcher(key)
	}
}

func (m *Middleware) MiddlewaresHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startTime := time.Now().Nanosecond()

		xReqId := r.Header.Get("X-Request-Id")
		if xReqId == "" {
			xReqId = uuid.NewV4().String()
		}
		r.Header.Set("X-Request-Id", xReqId)
		w.Header().Add("X-Request-Id", xReqId)

		token := r.Header.Get("Access-Token")
		if token != "" {
			r.Header.Set("Access-Token", token)
			w.Header().Add("Access-Token", token)
		}
		m.SetCorsHeaders(w, r)
		lrw := NewLoggingResponseWriter(w)
		h.ServeHTTP(lrw, r)
		IncRequestCount(r, lrw)
		SetRequestDuration(r, lrw, startTime)
	})
}

// Logger -
func (m *Middleware) LoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exceptions := make([]string, 0, 2)
		exceptions = append(exceptions, "kube-probe", "Prometheus")
		lrw := NewLoggingResponseWriter(w)
		config := zap.NewProductionConfig()
		config.Encoding = "json"
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		config.OutputPaths = []string{"stdout"}
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		config.EncoderConfig.CallerKey = "logger"
		logger, _ := config.Build()

		t1 := time.Now()
		defer func() {
			for _, v := range exceptions {
				if !strings.HasPrefix(r.Header.Get(userAgent), v) {
					logger.Info("",
						zap.Any("message", map[string]interface{}{
							"return-code": lrw.StatusCode,
							"http-method": r.Method,
							"request-headers": map[string]interface{}{
								"Content-Type":   r.Header.Get("Content-Type"),
								"Content-Length": r.Header.Get("Content-Length"),
								userAgent:        r.Header.Get(userAgent),
								"Server":         r.Header.Get("Server"),
								"Via":            r.Header.Get("Via"),
								"Accept":         r.Header.Get("Accept"),
								"pstxid":         r.Header.Get("pstxid"),
								"x-ps-pstxid":    r.Header.Get("x-ps-pstxid"),
								"x-ps-sso-token": r.Header.Get("x-ps-sso-token"),
							},
						}),

						zap.String("X-FORWARDED-FOR", r.Header.Get("X-FORWARDED-FOR")),
						zap.String("Remote Addr", r.RemoteAddr),
						zap.String("Proto", r.Proto),
						zap.String("Path", r.URL.Path),
						zap.Duration("Latency", time.Since(t1)),
					)
				}
			}
		}()
		h.ServeHTTP(lrw, r)
	})
}
