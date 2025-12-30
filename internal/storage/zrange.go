package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) ZRange(ctx context.Context, key []byte, start, stop int64) ([][]byte, error) {
	if hasError(ctxFlush(ctx)) {
		return nil, ErrContextCanceled
	}

	if isEmpty(key) {
		return [][]byte{}, nil
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return nil, err
	}

	var result [][]byte

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		if isNotFound(txnErr) {
			result = [][]byte{}
			return nil
		}

		if hasError(txnErr) {
			return txnErr
		}

		if !hasValidSortedSetHeader(data) {
			result = [][]byte{}
			return nil
		}

		count := int64(binary.LittleEndian.Uint64(data[:sortedSetHeaderSize]))

		if count == emptyCount {
			result = [][]byte{}
			return nil
		}

		if isNegativeIndex(start) {
			start = count + start
		}

		if isNegativeIndex(stop) {
			stop = count + stop
		}

		if start < firstElement {
			start = firstElement
		}

		if stop >= count {
			stop = count - singleItem
		}

		if start > stop {
			result = [][]byte{}
			return nil
		}

		members := make([][]byte, firstElement)
		offset := sortedSetHeaderSize

		for i := int64(firstElement); i < count; i++ {
			if offset+scoreSize > len(data) {
				break
			}

			offset += scoreSize

			if offset+itemLengthSize > len(data) {
				break
			}

			memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if offset+memberLen > len(data) {
				break
			}

			if i >= start && i <= stop {
				member := make([]byte, memberLen)
				copy(member, data[offset:offset+memberLen])
				members = append(members, member)
			}

			offset += memberLen
		}

		result = members
		return nil
	})

	if hasError(err) {
		return nil, err
	}

	return result, nil
}
