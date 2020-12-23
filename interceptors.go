package helpers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"time"
	"strings"
)

func GrpcServerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
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
		if !strings.HasPrefix(httpRequestUserAgent, v) {
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
			"response":     h,
			"time":         time.Since(start),
		}).Info("Response")
	}
	return h, err
}
