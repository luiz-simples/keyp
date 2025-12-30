package storage

import (
	"context"
)

func (client *Client) Incr(ctx context.Context, key []byte) (int64, error) {
	return client.IncrBy(ctx, key, defaultIncrement)
}
