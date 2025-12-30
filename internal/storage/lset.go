package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LSet(ctx context.Context, key []byte, index int64, value []byte) error {
	if hasError(ctxFlush(ctx)) {
		return ctx.Err()
	}

	if isEmpty(key) {
		return ErrKeyNotFound
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return ErrKeyNotFound
	}

	return client.env.Update(func(txn *lmdb.Txn) error {
		data, err := txn.Get(db, key)
		if hasError(err) {
			return ErrKeyNotFound
		}

		if len(data) < integerSize {
			return ErrKeyNotFound
		}

		length := int64(binary.LittleEndian.Uint64(data[:integerSize]))
		if index < firstElement {
			index = length + index
		}

		if index < firstElement || index >= length {
			return ErrKeyNotFound
		}

		newData := make([]byte, integerSize)
		binary.LittleEndian.PutUint64(newData, uint64(length))

		offset := integerSize
		for i := int64(firstElement); i < length; i++ {
			if offset+itemLengthSize > len(data) {
				return ErrKeyNotFound
			}

			itemLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if i == index {
				newData = append(newData, make([]byte, itemLengthSize)...)
				binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(len(value)))
				newData = append(newData, value...)
				offset += itemLen
			} else {
				if offset+itemLen > len(data) {
					return ErrKeyNotFound
				}
				newData = append(newData, make([]byte, itemLengthSize)...)
				binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(itemLen))
				newData = append(newData, data[offset:offset+itemLen]...)
				offset += itemLen
			}
		}

		return txn.Put(db, key, newData, noFlags)
	})
}
