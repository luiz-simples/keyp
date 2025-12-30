package storage

import (
	"context"
	"time"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (client *Client) Expire(ctx context.Context, keyBytes []byte, secs uint32) {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	db, _ := ctx.Value(domain.DB).(uint8)
	keys, hasKeys := client.ttl[db]

	if !hasKeys {
		client.ttl[db] = make(map[string]*TTL)
	}

	key := string(keyBytes)

	if ttl, hasTTL := keys[key]; hasTTL {
		ttl.Cancel()
	}

	client.ttl[db][key] = &TTL{
		Expire: uint32(time.Now().Unix()) + secs,
		Cancel: setTimeout(secs, func() {
			_, _ = client.Del(ctx, keyBytes)
		}),
	}
}
