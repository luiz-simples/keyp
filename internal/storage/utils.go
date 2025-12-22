package storage

import (
	"encoding/binary"
	"time"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func HasError(err error) bool {
	return err != nil
}

func IsEmpty(data any) bool {
	if byteSlice, ok := data.([]byte); ok {
		return len(byteSlice) == 0
	}

	if byteSliceSlice, ok := data.([][]byte); ok {
		return len(byteSliceSlice) == 0
	}

	return false
}

func exceedsLimit(key []byte) bool {
	return len(key) > MaxKeySize
}

func isNotFound(err error) bool {
	if lmdb.IsNotFound(err) {
		return true
	}
	return err == ErrKeyNotFound
}

func validateTTLKey(key []byte) error {
	if IsEmpty(key) {
		return ErrEmptyKey
	}

	if exceedsLimit(key) {
		return ErrKeyTooLarge
	}

	return nil
}

func validateTimestamp(timestamp int64) error {
	if isNegativeTimestamp(timestamp) {
		return ErrInvalidTimestamp
	}

	if isFutureTimestamp(timestamp) {
		return ErrInvalidTimestamp
	}

	return nil
}

func isNegativeTimestamp(timestamp int64) bool {
	return timestamp < 0
}

func isFutureTimestamp(timestamp int64) bool {
	maxFuture := time.Now().Unix() + (365 * 24 * 3600)
	return timestamp > maxFuture
}

func isExpiredBefore(expiresAt, before int64) bool {
	return expiresAt <= before
}

func isInvalidTTLData(data []byte) bool {
	return len(data) != TimestampSize*2
}

func calculateExpiresAt(seconds int64) int64 {
	return time.Now().Unix() + seconds
}

func calculateRemainingSeconds(expiresAt int64) int64 {
	return expiresAt - time.Now().Unix()
}

func calculateRemainingMilliseconds(expiresAt int64) int64 {
	remainingSeconds := expiresAt - time.Now().Unix()
	return remainingSeconds * 1000
}

func isNegativeSeconds(seconds int64) bool {
	return seconds < 0
}

func isPastTimestamp(timestamp int64) bool {
	return timestamp <= time.Now().Unix()
}

func isExpiredTime(remaining int64) bool {
	return remaining <= 0
}

func isKeyExpired(expiresAt int64) bool {
	return time.Now().Unix() >= expiresAt
}

func serializeTTLMetadata(metadata *TTLMetadata) []byte {
	result := make([]byte, TimestampSize*2)
	binary.BigEndian.PutUint64(result[0:TimestampSize], uint64(metadata.ExpiresAt))
	binary.BigEndian.PutUint64(result[TimestampSize:TimestampSize*2], uint64(metadata.CreatedAt))
	return result
}

func deserializeTTLMetadata(data []byte, key []byte) (*TTLMetadata, error) {
	if isInvalidTTLData(data) {
		return nil, ErrTTLCorrupted
	}

	expiresAt := int64(binary.BigEndian.Uint64(data[0:TimestampSize]))
	createdAt := int64(binary.BigEndian.Uint64(data[TimestampSize : TimestampSize*2]))

	keyCopy := make([]byte, len(key))
	copy(keyCopy, key)

	return &TTLMetadata{
		Key:       keyCopy,
		ExpiresAt: expiresAt,
		CreatedAt: createdAt,
	}, nil
}
