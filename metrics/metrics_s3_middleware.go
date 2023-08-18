package metrics

import (
	"context"
	"github.com/aws/smithy-go/middleware"
	"time"
)

func S3MetricsMiddleware(m Metrics) middleware.DeserializeMiddleware {
	metricsMiddleware := middleware.DeserializeMiddlewareFunc("ReportRequestMetrics", func(
		ctx context.Context, in middleware.DeserializeInput, next middleware.DeserializeHandler,
	) (
		out middleware.DeserializeOutput, metadata middleware.Metadata, err error,
	) {
		start := time.Now()
		out, metadata, err = next.HandleDeserialize(ctx, in)
		if err != nil {
			m.Inc(StorageRequestError.Name, "s3")
			return out, metadata, err
		}

		elapsed := time.Since(start).Seconds()

		m.Observe(StorageRequestExecutionTime.Name, elapsed, "s3")

		m.Inc(StorageRequestSuccessful.Name, "s3")

		return out, metadata, nil
	})

	return metricsMiddleware
}
