package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LIndex(ctx context.Context, key []byte, index int64) ([]byte, error) {
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

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		if hasError(txnErr) {
			return ErrKeyNotFound
		}

		if !hasValidListHeader(data) {
			return ErrKeyNotFound
		}

		length := int64(binary.LittleEndian.Uint64(data[:listHeaderSize]))

		if isNegativeIndex(index) {
			index = length + index
		}

		if isIndexOutOfBounds(index, length) {
			return ErrKeyNotFound
		}

		offset := listHeaderSize
		for i := int64(0); i <= index; i++ {
			if offset+itemLengthSize > len(data) {
				return ErrKeyNotFound
			}

			itemLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if i == index {
				if offset+itemLen > len(data) {
					return ErrKeyNotFound
				}
				result = make([]byte, itemLen)
				copy(result, data[offset:offset+itemLen])
				return nil
			}

			offset += itemLen
		}

		return ErrKeyNotFound
	})

	if hasError(err) {
		return nil, err
	}

	return result, nil
}
