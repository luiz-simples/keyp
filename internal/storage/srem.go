package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) SRem(ctx context.Context, key []byte, members ...[]byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) || isEmpty(members) {
		return emptyCount
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount
	}

	var removedCount int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)
		if hasError(txnErr) {
			return nil
		}

		if !hasValidSetHeader(data) {
			return nil
		}

		existingMembers := make(map[string]bool)
		count := int64(binary.LittleEndian.Uint64(data[:setHeaderSize]))
		offset := setHeaderSize

		for i := int64(firstElement); i < count; i++ {
			if offset+itemLengthSize > len(data) {
				break
			}

			memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
			offset += itemLengthSize

			if offset+memberLen > len(data) {
				break
			}

			member := string(data[offset : offset+memberLen])
			existingMembers[member] = true
			offset += memberLen
		}

		membersToRemove := make(map[string]bool)
		for _, member := range members {
			memberStr := string(member)
			if existingMembers[memberStr] {
				membersToRemove[memberStr] = true
				removedCount++
			}
		}

		if removedCount == emptyCount {
			return nil
		}

		for memberStr := range membersToRemove {
			delete(existingMembers, memberStr)
		}

		if len(existingMembers) == emptyCount {
			return txn.Del(db, key, nil)
		}

		newData := make([]byte, setHeaderSize)
		newCount := int64(len(existingMembers))
		binary.LittleEndian.PutUint64(newData[:setHeaderSize], uint64(newCount))

		for memberStr := range existingMembers {
			memberBytes := []byte(memberStr)
			newData = append(newData, make([]byte, itemLengthSize)...)
			binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(len(memberBytes)))
			newData = append(newData, memberBytes...)
		}

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return emptyCount
	}

	return removedCount
}
