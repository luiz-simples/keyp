package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) SMembers(ctx context.Context, key []byte) ([][]byte, error) {
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

		if !hasValidSetHeader(data) {
			return nil
		}

		count := int64(binary.LittleEndian.Uint64(data[:setHeaderSize]))
		offset := setHeaderSize

		for i := int64(0); i < count; i++ {
			if offset+itemLengthSize > len(data) {
				return ErrKeyNotFound
			}

			memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if offset+memberLen > len(data) {
				return ErrKeyNotFound
			}

			member := make([]byte, memberLen)
			copy(member, data[offset:offset+memberLen])
			result = append(result, member)
			offset += memberLen
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
