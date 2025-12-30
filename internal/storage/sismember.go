package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) SIsMember(ctx context.Context, key []byte, member []byte) bool {
	if hasError(ctxFlush(ctx)) {
		return false
	}

	if isEmpty(key) || isEmpty(member) {
		return false
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return false
	}

	var found bool

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
		memberStr := string(member)

		for i := int64(firstElement); i < count; i++ {
			if offset+itemLengthSize > len(data) {
				return nil
			}

			memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if offset+memberLen > len(data) {
				return nil
			}

			existingMember := string(data[offset : offset+memberLen])
			if existingMember == memberStr {
				found = true
				return nil
			}

			offset += memberLen
		}

		return nil
	})

	if hasError(err) {
		return false
	}

	return found
}
