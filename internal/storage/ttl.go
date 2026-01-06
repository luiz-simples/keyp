package storage

import (
	"context"
	"time"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (client *Client) TTL(ctx context.Context, keyBytes []byte) uint32 {
	client.mtx.RLock()
	defer client.mtx.RUnlock()

	db, _ := ctx.Value(domain.DB).(uint8)
	key := string(keyBytes)
	keys, hasKeys := client.ttl[db]

	if !hasKeys {
		return 0
	}

	ttl := keys[key]

	if ttl == nil {
		return 0
	}

	now := uint32(time.Now().Unix())
	if ttl.Expire <= now {
		return 0
	}

	return ttl.Expire - now
}
