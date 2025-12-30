package storage

import (
	"context"
	"encoding/binary"
	"math"
	"sort"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

type scoredMember struct {
	score  float64
	member string
}

func (client *Client) ZAdd(ctx context.Context, key []byte, score float64, member []byte) int64 {
	if hasError(ctxFlush(ctx)) {
		return emptyCount
	}

	if isEmpty(key) || isEmpty(member) {
		return emptyCount
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount
	}

	var addedCount int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		existingMembers := make(map[string]float64)

		if isEmpty(txnErr) && hasValidSortedSetHeader(data) {
			count := int64(binary.LittleEndian.Uint64(data[:sortedSetHeaderSize]))
			offset := sortedSetHeaderSize

			for i := int64(firstElement); i < count; i++ {
				if offset+scoreSize > len(data) {
					break
				}

				memberScore := math.Float64frombits(binary.LittleEndian.Uint64(data[offset:]))
				offset += scoreSize

				if offset+itemLengthSize > len(data) {
					break
				}

				memberLen := int(binary.LittleEndian.Uint32(data[offset:]))
				offset += itemLengthSize

				if offset+memberLen > len(data) {
					break
				}

				memberStr := string(data[offset : offset+memberLen])
				existingMembers[memberStr] = memberScore
				offset += memberLen
			}
		}

		memberStr := string(member)
		if _, exists := existingMembers[memberStr]; !exists {
			addedCount = singleItem
		}

		existingMembers[memberStr] = score

		scoredMembers := make([]scoredMember, firstElement, len(existingMembers))
		for memberStr, memberScore := range existingMembers {
			scoredMembers = append(scoredMembers, scoredMember{
				score:  memberScore,
				member: memberStr,
			})
		}

		sort.Slice(scoredMembers, func(i, j int) bool {
			if scoredMembers[i].score == scoredMembers[j].score {
				return scoredMembers[i].member < scoredMembers[j].member
			}
			return scoredMembers[i].score < scoredMembers[j].score
		})

		newData := make([]byte, sortedSetHeaderSize)
		binary.LittleEndian.PutUint64(newData[:sortedSetHeaderSize], uint64(len(scoredMembers)))

		for _, sm := range scoredMembers {
			newData = append(newData, make([]byte, scoreSize)...)
			binary.LittleEndian.PutUint64(newData[len(newData)-scoreSize:], math.Float64bits(sm.score))

			memberBytes := []byte(sm.member)
			newData = append(newData, make([]byte, itemLengthSize)...)
			binary.LittleEndian.PutUint32(newData[len(newData)-itemLengthSize:], uint32(len(memberBytes)))
			newData = append(newData, memberBytes...)
		}

		return txn.Put(db, key, newData, noFlags)
	})

	if hasError(err) {
		return emptyCount
	}

	return addedCount
}
