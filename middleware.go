package helpers

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	uuid "github.com/satori/go.uuid"
	"google.golang.org/protobuf/runtime/protoiface"
	"net/http"
	"time"
)

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
