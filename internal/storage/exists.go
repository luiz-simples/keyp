package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) Exists(ctx context.Context, key []byte) bool {
	if hasError(ctxFlush(ctx)) {
		return false
	}

	if isEmpty(key) {
		return false
	}

	db, err := client.sel(ctx)

	if hasError(err) {
		return false
	}

	err = client.env.View(func(txn *lmdb.Txn) error {
		_, txnErr := txn.Get(db, key)
		return txnErr
	})

	return isEmpty(err)
}
