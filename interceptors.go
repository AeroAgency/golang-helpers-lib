package helpers

import (
	"context"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"os"
	"time"
)

func GrpcServerUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	md, _ := metadata.FromIncomingContext(ctx)
	var xRequestId string
	for i, v := range md {
		if i == "x-request-id" {
			xRequestId = v[0]
		}
	}
	logger := log.Logger{}
	logger.SetFormatter(&log.JSONFormatter{})
	logger.SetOutput(os.Stdout)
	logger.Info("run grpc server")
	logger.SetLevel(log.DebugLevel)
	logger.WithFields(log.Fields{
		"x-request-id": xRequestId,
		"method":       info.FullMethod,
		//"request":      req,
	}).Info("Request")

	// Calls the handler
	h, err := handler(ctx, req)

	logger.WithFields(log.Fields{
		"x-request-id": xRequestId,
		"method":       info.FullMethod,
		"response":     h,
		"time":         time.Since(start),
	}).Info("Response")

	return h, err
}
