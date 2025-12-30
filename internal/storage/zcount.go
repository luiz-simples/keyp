package storage

import (
	"context"
	"encoding/binary"
	"math"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) ZCount(ctx context.Context, key []byte, min, max float64) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) {
		return emptyCount
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount
	}

	var count int64

	err = client.env.View(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		if isNotFound(txnErr) {
			count = emptyCount
			return nil
		}

		if hasError(txnErr) {
			return txnErr
		}

		if !hasValidSortedSetHeader(data) {
			count = emptyCount
			return nil
		}

		totalCount := int64(binary.LittleEndian.Uint64(data[:sortedSetHeaderSize]))

		if totalCount == emptyCount {
			count = emptyCount
			return nil
		}

		offset := sortedSetHeaderSize

		for i := int64(firstElement); i < totalCount; i++ {
			if offset+scoreSize > len(data) {
				break
			}

			score := math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
			offset += scoreSize

			if offset+itemLengthSize > len(data) {
				break
			}

			memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if offset+memberLen > len(data) {
				break
			}

			if isScoreInRange(score, min, max) {
				count++
			}

			offset += memberLen
		}

		return nil
	})

	if hasError(err) {
		return emptyCount
	}

	return count
}
