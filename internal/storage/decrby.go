package storage

import (
	"context"
	"strconv"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

func (client *Client) DecrBy(ctx context.Context, key []byte, decrement int64) (int64, error) {
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

		var currentValue int64

		if isNotFound(txnErr) {
			currentValue = emptyCount
		} else if hasError(txnErr) {
			return txnErr
		} else {
			parsedValue, parseErr := strconv.ParseInt(string(data), 10, 64)
			if hasError(parseErr) {
				return ErrNotInteger
			}
			currentValue = parsedValue
		}

		result = currentValue - decrement
		newValue := strconv.FormatInt(result, 10)

		return txn.Put(db, key, []byte(newValue), noFlags)
	})

	if hasError(err) {
		return emptyCount, err
	}

	return result, nil
}
