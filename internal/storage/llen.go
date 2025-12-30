package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LLen(ctx context.Context, key []byte) int64 {
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

	var length int64

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)
		if hasError(txnErr) {
			return txnErr
		}

		if len(data) < integerSize {
			return nil
		}

		length = int64(binary.LittleEndian.Uint64(data[:integerSize]))
		return nil
	})

	if hasError(err) {
		return emptyCount
	}

	return length
}
