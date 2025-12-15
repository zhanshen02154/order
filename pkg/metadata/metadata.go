package metadata

import (
	"context"
	"go-micro.dev/v4/metadata"
)

func GetValueFromMetadata(ctx context.Context, key string) string {
	val, ok := metadata.Get(ctx, key)
	if !ok {
		return ""
	}
	return val
}
