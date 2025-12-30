package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LPop(ctx context.Context, key []byte) ([]byte, error) {
	if hasError(ctxFlush(ctx)) {
		return nil, ctx.Err()
	}

	if isEmpty(key) {
		return nil, ErrKeyNotFound
	}

	db, err := client.sel(ctx)

	if hasError(err) {
		return nil, ErrKeyNotFound
	}

	var result []byte

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		if hasError(txnErr) {
			return ErrKeyNotFound
		}

		if len(data) < integerSize {
			return ErrKeyNotFound
		}

		length := int64(binary.LittleEndian.Uint64(data[:integerSize]))

		if length == emptyCount {
			return ErrKeyNotFound
		}

		if length == singleItem {
			if len(data) < integerSize+itemLengthSize {
				return ErrKeyNotFound
			}
			itemLen := int(binary.LittleEndian.Uint32(data[integerSize : integerSize+itemLengthSize]))
			if len(data) < integerSize+itemLengthSize+itemLen {
				return ErrKeyNotFound
			}
			result = make([]byte, itemLen)
			copy(result, data[integerSize+itemLengthSize:integerSize+itemLengthSize+itemLen])
			return txn.Del(db, key, nil)
		}

		if len(data) < integerSize+itemLengthSize {
			return ErrKeyNotFound
		}

		firstItemLen := int(binary.LittleEndian.Uint32(data[integerSize : integerSize+itemLengthSize]))
		if len(data) < integerSize+itemLengthSize+firstItemLen {
			return ErrKeyNotFound
		}

		result = make([]byte, firstItemLen)
		copy(result, data[integerSize+itemLengthSize:integerSize+itemLengthSize+firstItemLen])

		newData := make([]byte, integerSize)
		binary.LittleEndian.PutUint64(newData, uint64(length-singleItem))
		newData = append(newData, data[integerSize+itemLengthSize+firstItemLen:]...)

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return nil, err
	}

	return result, nil
}
