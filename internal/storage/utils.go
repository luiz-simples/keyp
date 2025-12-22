package storage

import "github.com/PowerDNS/lmdb-go/lmdb"

func hasError(err error) bool {
	return err != nil
}

func isEmpty(key []byte) bool {
	return len(key) == 0
}

func exceedsLimit(key []byte) bool {
	return len(key) > MaxKeySize
}

func isNotFound(err error) bool {
	return lmdb.IsNotFound(err)
}
