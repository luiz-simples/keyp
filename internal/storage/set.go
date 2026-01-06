package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) Set(ctx context.Context, key, val []byte) error {
	db, err := client.sel(ctx)

	if noError(err) {
		err = ctxFlush(ctx)
	}

	if hasError(err) {
		return err
	}

	return client.env.Update(func(txn *lmdb.Txn) error {
		if err := ctxFlush(ctx); hasError(err) {
			return err
		}

		return txn.Put(db, key, val, noFlags)
	})
}
