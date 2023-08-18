package metrics

import (
	"context"
	goRedis "github.com/go-redis/redis/v8"
	"time"
)

type startKey struct{}

type RedisHook struct {
	metrics Metrics
}

func NewRedisHook(metrics Metrics) *RedisHook {
	return &RedisHook{metrics: metrics}
}

func (h RedisHook) BeforeProcess(ctx context.Context, cmd goRedis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (h RedisHook) AfterProcess(ctx context.Context, cmd goRedis.Cmder) error {
	if start, ok := ctx.Value(startKey{}).(time.Time); ok {
		duration := time.Since(start).Seconds()
		h.metrics.Observe(StorageRequestExecutionTime.Name, duration, "redis")
	}

	metric := StorageRequestSuccessful.Name
	if cmd.Err() != nil && cmd.Err() != goRedis.Nil {
		metric = StorageRequestError.Name
	}
	h.metrics.Inc(metric, "redis")

	return nil
}

func (h RedisHook) BeforeProcessPipeline(ctx context.Context, cmds []goRedis.Cmder) (context.Context, error) {
	return context.WithValue(ctx, startKey{}, time.Now()), nil
}

func (h RedisHook) AfterProcessPipeline(ctx context.Context, cmds []goRedis.Cmder) error {
	if err := h.AfterProcess(ctx, goRedis.NewCmd(ctx, "pipeline")); err != nil {
		return err
	}

	for _, cmd := range cmds {
		metric := StorageRequestSuccessful.Name
		if cmd.Err() != nil && cmd.Err() != goRedis.Nil {
			metric = StorageRequestError.Name
		}
		h.metrics.Inc(metric, "redis")
	}

	return nil
}
