package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/common/log"
	uuid "github.com/satori/go.uuid"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/runtime/protoiface"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"strings"
	"time"
)

// userAgent - duplicate string in code
const userAgent = "User-Agent"

const defaultErrorMessage = "Сервер не смог обработать ваш запрос. Проверьте корректность типов входных параметров"

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
		var exception bool
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
				if strings.HasPrefix(r.Header.Get(userAgent), v) {
					exception = true
				}
			}
			if !exception {
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
							"pstxid":         r.Header.Get("X-Request-Id"),
							"x-ps-sso-token": r.Header.Get("Access-Token"),
						},
					}),

					zap.String("X-FORWARDED-FOR", r.Header.Get("X-FORWARDED-FOR")),
					zap.String("Remote Addr", r.RemoteAddr),
					zap.String("Proto", r.Proto),
					zap.String("Path", r.URL.Path),
					zap.Duration("Latency", time.Since(t1)),
				)
			}
		}()
		h.ServeHTTP(lrw, r)
	})
}

// Кастомная функция для отображения ошибок, пришедших от микросервисов по gRPC
func (m *Middleware) CustomHTTPError(ctx context.Context, _ *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, _ *http.Request, err error) {
	const fallback = `{"error": "failed to marshal error message"}`
	var errorsDataMessage string
	type errorBody struct {
		Err     string `json:"error"`
		Message string `json:"message"`
	}
	w.Header().Set("Content-type", marshaler.ContentType())
	w.WriteHeader(runtime.HTTPStatusFromCode(status.Code(err)))

	st := status.Convert(err)
	for _, detail := range st.Details() {
		switch t := detail.(type) {
		case *errdetails.ErrorInfo:
			errorsDataMessage = t.Metadata["message"]
		}
	}
	if errorsDataMessage == "" {
		errorsDataMessage = defaultErrorMessage
	}
	jErr := json.NewEncoder(w).Encode(errorBody{
		Err:     status.Convert(err).Message(),
		Message: errorsDataMessage,
	})
	if jErr != nil {
		w.Write([]byte(fallback))
	}
}

// Логирует REST запросы к серверу
func (m *Middleware) RestRequestLoggerMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpRequestUserAgent := "User-Agent"
		var exception bool
		exceptions := make([]string, 0, 2)
		exceptions = append(exceptions, "kube-probe", "Prometheus")
		restLogmode, err := strconv.Atoi(os.Getenv("REST_LOGMODE"))
		if err == nil && restLogmode == 1 {
			requestDump, err := httputil.DumpRequest(r, true)
			for _, v := range exceptions {
				if strings.HasPrefix(r.Header.Get(httpRequestUserAgent), v) {
					exception = true
				}
			}
			if err == nil && exception != true {
				log.Info("REST REQUEST")
				fmt.Print(string(requestDump))

			}
		}
		h.ServeHTTP(w, r)
	})
}

// Middleware для вывода тела ответа
func RestResponseLoggerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	md, _ := metadata.FromIncomingContext(ctx)
	var httpRequestUserAgent string
	var exception bool
	exceptions := make([]string, 0, 2)
	exceptions = append(exceptions, "kube-probe", "Prometheus")
	for i, v := range md {
		if i == "grpcgateway-user-agent" {
			httpRequestUserAgent = v[0]
		}
	}
	for _, v := range exceptions {
		if strings.HasPrefix(httpRequestUserAgent, v) {
			exception = true
		}
	}
	// Calls the handler
	h, err := handler(ctx, req)
	if exception != true {
		log.Info(fmt.Sprintf("REST RESPONSE CODE %d", runtime.HTTPStatusFromCode(status.Code(err))))
		log.Info("REST RESPONSE BODY")
		if err == nil {
			b, _ := json.MarshalIndent(h, "", "    ")
			log.Info(string(b))
		} else {
			st := status.Convert(err)
			for _, detail := range st.Details() {
				switch t := detail.(type) {
				case *errdetails.ErrorInfo:
					type errorBody struct {
						Err     string `json:"error"`
						Message string `json:"message"`
					}
					errorData := errorBody{
						Err:     st.Message(),
						Message: t.Metadata["message"],
					}
					b, _ := json.MarshalIndent(errorData, "", "    ")
					fmt.Println(string(b))
				}
			}
		}
	}
	return h, err
}
