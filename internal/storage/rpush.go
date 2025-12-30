package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) RPush(ctx context.Context, key []byte, values ...[]byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return 0
	}

	if isEmpty(key) || isEmpty(values) {
		return 0
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return 0
	}

	var newLength int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		var currentLength int64
		newData := make([]byte, 8)

		if isEmpty(txnErr) && len(data) >= 8 {
			currentLength = int64(binary.LittleEndian.Uint64(data[:8]))
			newData = append(newData, data[8:]...)
		}

		newLength = currentLength + int64(len(values))
		binary.LittleEndian.PutUint64(newData[:8], uint64(newLength))

		for _, value := range values {
			newData = append(newData, make([]byte, 4)...)
			binary.LittleEndian.PutUint32(newData[len(newData)-4:], uint32(len(value)))
			newData = append(newData, value...)
		}

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return 0
	}

	return newLength
}
