package storage

import (
	"context"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) Append(ctx context.Context, key, value []byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) {
		return emptyCount
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount
	}

	var newLength int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		var newData []byte

		if isNotFound(txnErr) {
			newData = make([]byte, len(value))
			copy(newData, value)
		} else if hasError(txnErr) {
			return txnErr
		} else {
			newData = make([]byte, len(data)+len(value))
			copy(newData, data)
			copy(newData[len(data):], value)
		}

		newLength = int64(len(newData))
		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return emptyCount
	}

	return newLength
}
