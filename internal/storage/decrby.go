package storage

import (
	"context"
)

func (client *Client) DecrBy(ctx context.Context, key []byte, decrement int64) (int64, error) {
	return modifyIntegerBy(client, ctx, key, decrement, func(current, delta int64) int64 {
		return current - delta
	})
}
