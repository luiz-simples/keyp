package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) FlushAll(ctx context.Context) error {
	if hasError(ctxFlush(ctx)) {
		return ctx.Err()
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return err
	}

	return client.env.Update(func(txn *lmdb.Txn) error {
		cursor, cursorErr := txn.OpenCursor(db)
		if hasError(cursorErr) {
			return cursorErr
		}
		defer cursor.Close()

		for {
			_, _, cursorErr := cursor.Get(nil, nil, lmdb.First)
			if hasError(cursorErr) {
				if lmdb.IsNotFound(cursorErr) {
					return nil
				}
				return cursorErr
			}

			delErr := cursor.Del(noFlags)
			if hasError(delErr) {
				return delErr
			}
		}
	})
}
