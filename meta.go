package helpers

import (
	"context"
	b64 "encoding/base64"
	"google.golang.org/grpc/metadata"
)

type Meta struct{}

func (m Meta) GetParam(ctx context.Context, key string) string {
	md, _ := metadata.FromIncomingContext(ctx)
	var value string
	for i, v := range md {
		if i == key {
			value = v[0]
		}
	}
	return value
}

func (m Meta) GetDecodedParam(ctx context.Context, key string) (string, error) {
	value := m.GetParam(ctx, key)
	decodedValue, err := b64.StdEncoding.DecodeString(value)
	if err != nil {
		return "", err
	}
	return string(decodedValue), nil
}
