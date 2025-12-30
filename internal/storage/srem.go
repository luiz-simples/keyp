package storage

import (
	"context"
	"encoding/binary"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) SRem(ctx context.Context, key []byte, members ...[]byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return 0
	}

	if isEmpty(key) || isEmpty(members) {
		return 0
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return 0
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

		for i := int64(0); i < count; i++ {
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

		if removedCount == 0 {
			return nil
		}

		for memberStr := range membersToRemove {
			delete(existingMembers, memberStr)
		}

		if len(existingMembers) == 0 {
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
		return 0
	}

	return removedCount
}
