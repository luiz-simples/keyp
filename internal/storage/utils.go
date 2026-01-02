package storage

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

const (
	listHeaderSize      = 8
	setHeaderSize       = 8
	sortedSetHeaderSize = 8
	itemLengthSize      = 4
	scoreSize           = 8

	integerSize = 8

	firstElement = 0
	emptyCount   = 0
	singleItem   = 1

	defaultIncrement = 1
	defaultDecrement = -1
)

func hasError(err error) bool {
	return err != nil
}

func isEmpty(data any) bool {
	str, isStr := data.(string)

	if isStr {
		return strings.TrimSpace(str) == ""
	}

	bytes, isBytes := data.([]byte)

	if isBytes {
		return len(bytes) == 0
	}

	listBytes, isListBytes := data.([][]byte)

	if isListBytes {
		return len(listBytes) == 0
	}

	return data == nil
}

func isNotFound(err error) bool {
	if isEmpty(err) {
		return false
	}

	if lmdb.IsNotFound(err) {
		return true
	}

	return err == ErrKeyNotFound
}

func hasValidListHeader(data []byte) bool {
	return len(data) >= listHeaderSize
}

func hasValidSetHeader(data []byte) bool {
	return len(data) >= setHeaderSize
}

func hasValidSortedSetHeader(data []byte) bool {
	return len(data) >= sortedSetHeaderSize
}

func isScoreInRange(score, min, max float64) bool {
	return score >= min && score <= max
}

func isNegativeIndex(index int64) bool {
	return index < 0
}

func isIndexOutOfBounds(index, length int64) bool {
	return index < 0 || index >= length
}

func hasInvalidListHeader(data []byte) bool {
	return len(data) < integerSize
}

func hasInsufficientData(data []byte, offset, size int) bool {
	return offset+size > len(data)
}

func ctxFlush(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}

func setTimeout(secs uint32, fn func()) func() {
	running := true

	time.AfterFunc(time.Duration(secs)*time.Second, func() {
		if running {
			fn()
		}
	})

	return func() {
		running = false
	}
}

func modifyIntegerBy(client *Client, ctx context.Context, key []byte, delta int64, operation func(int64, int64) int64) (int64, error) {
	if hasError(ctxFlush(ctx)) {
		return emptyCount, ErrContextCanceled
	}

	if isEmpty(key) {
		return emptyCount, ErrKeyNotFound
	}

	db, err := client.sel(ctx)
	if hasError(err) {
		return emptyCount, err
	}

	var result int64

	err = client.env.Update(func(txn *lmdb.Txn) error {
		data, txnErr := txn.Get(db, key)

		if isNotFound(txnErr) {
			result = operation(emptyCount, delta)
			newValue := strconv.FormatInt(result, 10)
			return txn.Put(db, key, []byte(newValue), noFlags)
		}

		if hasError(txnErr) {
			return txnErr
		}

		parsedValue, parseErr := strconv.ParseInt(string(data), 10, 64)
		if hasError(parseErr) {
			return ErrNotInteger
		}

		result = operation(parsedValue, delta)
		newValue := strconv.FormatInt(result, 10)
		return txn.Put(db, key, []byte(newValue), noFlags)
	})

	if hasError(err) {
		return emptyCount, err
	}

	return result, nil
}
