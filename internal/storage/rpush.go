package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) RPush(ctx context.Context, key []byte, values ...[]byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) || isEmpty(values) {
		return emptyCount
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount
	}

	var newLength int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		var currentLength int64
		newData := make([]byte, integerSize)

		if isEmpty(txnErr) && len(data) >= integerSize {
			currentLength = int64(binary.LittleEndian.Uint64(data[:integerSize]))
			newData = append(newData, data[integerSize:]...)
		}

		newLength = currentLength + int64(len(values))
		binary.LittleEndian.PutUint64(newData[:integerSize], uint64(newLength))

		for _, value := range values {
			newData = append(newData, make([]byte, itemLengthSize)...)
			binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(len(value)))
			newData = append(newData, value...)
		}

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return emptyCount
	}

	return newLength
}
