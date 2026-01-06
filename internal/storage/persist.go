package storage

import (
	"context"

	"github.com/luiz-simples/keyp.git/internal/domain"
)

func (client *Client) Persist(ctx context.Context, keyBytes []byte) bool {
	client.mtx.Lock()
	defer client.mtx.Unlock()

	db, _ := ctx.Value(domain.DB).(uint8)
	key := string(keyBytes)
	keys, hasKeys := client.ttl[db]

	if !hasKeys {
		return false
	}

	ttl := keys[key]

	if ttl == nil {
		return false
	}

	ttl.Cancel()
	delete(keys, key)
	return true
}
