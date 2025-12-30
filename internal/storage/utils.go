package storage

import (
	"context"
	"time"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func hasError(err error) bool {
	return err != nil
}

func noError(err error) bool {
	return err == nil
}

func isEmpty(data any) bool {
	if byteSlice, ok := data.([]byte); ok {
		return len(byteSlice) == 0
	}

	if byteSliceSlice, ok := data.([][]byte); ok {
		return len(byteSliceSlice) == 0
	}

	return false
}

func isNotFound(err error) bool {
	if noError(err) {
		return false
	}

	if lmdb.IsNotFound(err) {
		return true
	}

	return err == ErrKeyNotFound
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
