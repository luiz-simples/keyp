package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) LRange(ctx context.Context, key []byte, start, stop int64) ([][]byte, error) {
	if hasError(ctxFlush(ctx)) {
		return nil, ctx.Err()
	}

	if isEmpty(key) {
		return [][]byte{}, nil
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return [][]byte{}, nil
	}

	var result [][]byte

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)
		if hasError(txnErr) {
			return nil
		}

		if len(data) < 8 {
			return nil
		}

		length := int64(binary.LittleEndian.Uint64(data[:8]))
		if length == 0 {
			return nil
		}

		if start < 0 {
			start = length + start
		}
		if stop < 0 {
			stop = length + stop
		}

		if start < 0 {
			start = 0
		}
		if stop >= length {
			stop = length - 1
		}

		if start > stop {
			return nil
		}

		offset := 8
		for i := int64(0); i < length; i++ {
			if offset+4 > len(data) {
				return ErrKeyNotFound
			}

			itemLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += 4

			if i >= start && i <= stop {
				if offset+itemLen > len(data) {
					return ErrKeyNotFound
				}
				item := make([]byte, itemLen)
				copy(item, data[offset:offset+itemLen])
				result = append(result, item)
			}

			offset += itemLen
		}

		return nil
	})

	if hasError(err) {
		return nil, err
	}

	if result == nil {
		result = [][]byte{}
	}

	return result, nil
}
