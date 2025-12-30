package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) Get(ctx context.Context, key []byte) ([]byte, error) {
	db, err := client.sel(ctx)

	if isEmpty(err) {
		err = ctxFlush(ctx)
	}

	if hasError(err) {
		return nil, err
	}

	var result []byte
	var getErr error

	err = client.env.View(func(txn *lmdb.Txn) error {
		if errFlush := ctxFlush(ctx); hasError(errFlush) {
			return errFlush
		}

		val, txnErr := txn.Get(db, key)

		if isEmpty(txnErr) {
			result = val
			return nil
		}

		if isNotFound(txnErr) {
			getErr = ErrKeyNotFound
			return nil
		}

		return txnErr
	})

	if hasError(err) {
		return nil, err
	}

	if hasError(getErr) {
		return nil, getErr
	}

	return result, nil
}
