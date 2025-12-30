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

		if len(data) < 8 {
			return ErrKeyNotFound
		}

		length := int64(binary.LittleEndian.Uint64(data[:8]))

		if length == 0 {
			return ErrKeyNotFound
		}

		if length == 1 {
			if len(data) < 12 {
				return ErrKeyNotFound
			}
			itemLen := int(binary.LittleEndian.Uint32(data[8:12]))
			if len(data) < 12+itemLen {
				return ErrKeyNotFound
			}
			result = make([]byte, itemLen)
			copy(result, data[12:12+itemLen])
			return txn.Del(db, key, nil)
		}

		if len(data) < 12 {
			return ErrKeyNotFound
		}

		firstItemLen := int(binary.LittleEndian.Uint32(data[8:12]))
		if len(data) < 12+firstItemLen {
			return ErrKeyNotFound
		}

		result = make([]byte, firstItemLen)
		copy(result, data[12:12+firstItemLen])

		newData := make([]byte, 8)
		binary.LittleEndian.PutUint64(newData, uint64(length-1))
		newData = append(newData, data[12+firstItemLen:]...)

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return nil, err
	}

	return result, nil
}
