package storage

import (
	"context"
)

func (client *Client) IncrBy(ctx context.Context, key []byte, increment int64) (int64, error) {
	return modifyIntegerBy(client, ctx, key, increment, func(current, delta int64) int64 {
		return current + delta
	})
}
