package storage

import (
	"context"
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
