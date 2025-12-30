package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

const EMPTY = uint32(0)

func (client *Client) Del(ctx context.Context, keys ...[]byte) (uint32, error) {
	db, err := client.sel(ctx)

	if isEmpty(err) {
		err = ctxFlush(ctx)
	}

	if hasError(err) {
		return EMPTY, err
	}

	if isEmpty(keys) {
		return EMPTY, nil
	}

	deleted := EMPTY

	err = client.env.Update(func(txn *lmdb.Txn) error {
		if errFlush := ctxFlush(ctx); hasError(errFlush) {
			return errFlush
		}

		for _, key := range keys {
			delErr := txn.Del(db, key, nil)

			if isEmpty(delErr) {
				deleted++
				continue
			}

			if isNotFound(delErr) {
				continue
			}

			return delErr
		}

		return nil
	})

	if hasError(err) {
		return EMPTY, err
	}

	return deleted, nil
}
