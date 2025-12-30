package storage

import (
	"context"
)

func (client *Client) Decr(ctx context.Context, key []byte) (int64, error) {
	return client.DecrBy(ctx, key, defaultIncrement)
}
