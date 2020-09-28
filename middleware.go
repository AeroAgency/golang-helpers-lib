package helpers

import (
	"context"
	"google.golang.org/protobuf/runtime/protoiface"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"net/http"
)

type Middleware struct {}

// Установка заголовков для CORS-запросов
func(m *Middleware) setCorsHeaders(w http.ResponseWriter, r *http.Request) {
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
func(m *Middleware) HttpResponseModifier(ctx context.Context, w http.ResponseWriter, proto protoiface.MessageV1) error {
	// очистка ненужных заголовков
	delete(w.Header(), "Grpc-Metadata-Content-Type")
	return nil
}

// Мэппинг для заголовков
func(m *Middleware) CustomMatcher(key string) (string, bool) {
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