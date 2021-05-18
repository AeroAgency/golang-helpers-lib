package helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	log "github.com/sirupsen/logrus"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"os"
	"strconv"
	"strings"
	"time"
)

// Middleware для вывода тела ответа REST
func RestResponseLoggerInterceptor(ctx context.Context, req interface{}, _ *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	restLogMode, _ := strconv.Atoi(os.Getenv("REST_LOGMODE"))
	if restLogMode != 1 {
		return handler(ctx, req)
	}
	md, _ := metadata.FromIncomingContext(ctx)
	var httpRequestUserAgent string
	var exception bool
	exceptions := getLogExceptions()
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
			fmt.Println(string(b))
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

// Middleware для вывода тела ответа GRPC
func GrpcServerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	logMode, _ := strconv.Atoi(os.Getenv("GRPC_LOGMODE"))
	if logMode != 1 {
		return handler(ctx, req)
	}
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	var httpRequestUserAgent string
	var exception bool
	exceptions := make([]string, 0, 2)
	exceptions = append(exceptions, "kube-probe", "Prometheus")
	var xRequestId string
	for i, v := range md {
		if i == "x-request-id" {
			xRequestId = v[0]
		}
		if i == "grpcgateway-user-agent" {
			httpRequestUserAgent = v[0]
		}
	}
	for _, v := range exceptions {
		if strings.HasPrefix(httpRequestUserAgent, v) {
			exception = true
		}
	}
	logger := log.Logger{}
	if exception != true {
		logger.SetFormatter(&log.JSONFormatter{})
		logger.SetOutput(os.Stdout)
		logger.Info("run grpc server")
		logger.SetLevel(log.DebugLevel)
		logger.WithFields(log.Fields{
			"x-request-id": xRequestId,
			"method":       info.FullMethod,
			//"request":      req,
		}).Info("Request")
	}

	// Calls the handler
	h, err := handler(ctx, req)
	if exception != true {
		logger.WithFields(log.Fields{
			"x-request-id": xRequestId,
			"method":       info.FullMethod,
			"error":        err,
			"response":     h,
			"time":         time.Since(start),
		}).Info("Response")
	}
	return h, err
}
