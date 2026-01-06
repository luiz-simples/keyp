package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LPush(ctx context.Context, key []byte, values ...[]byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) || isEmpty(values) {
		return emptyCount
	}

	typeErr := client.checkKeyType(ctx, key, "list")
	if hasError(typeErr) {
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
		var existingData []byte

		if isEmpty(txnErr) && len(data) >= integerSize {
			currentLength = int64(binary.LittleEndian.Uint64(data[:integerSize]))
			existingData = data[integerSize:]
		}

		newLength = currentLength + int64(len(values))
		newData := make([]byte, integerSize)
		binary.LittleEndian.PutUint64(newData, uint64(newLength))

		for i := len(values) - singleItem; i >= firstElement; i-- {
			newData = append(newData, make([]byte, itemLengthSize)...)
			binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(len(values[i])))
			newData = append(newData, values[i]...)
		}

		newData = append(newData, existingData...)

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return emptyCount
	}

	return newLength
}
