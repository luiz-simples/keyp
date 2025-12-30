package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LLen(ctx context.Context, key []byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return 0
	}

	if isEmpty(key) {
		return 0
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return 0
	}

	var length int64

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)
		if hasError(txnErr) {
			return txnErr
		}

		if len(data) < 8 {
			return nil
		}

		length = int64(binary.LittleEndian.Uint64(data[:8]))
		return nil
	})

	if hasError(err) {
		return 0
	}

	return length
}
