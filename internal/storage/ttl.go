package storage

import (
	"errors"
	"time"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

var (
	ErrTTLNotFound      = errors.New("TTL not found for key")
	ErrTTLCorrupted     = errors.New("TTL metadata corrupted")
	ErrInvalidTTL       = errors.New("invalid TTL value")
	ErrInvalidTimestamp = errors.New("invalid timestamp")
)

const (
	TTLDatabaseName = "keyp_ttl"
	TimestampSize   = 8
)

type TTLMetadata struct {
	Key       []byte
	ExpiresAt int64
	CreatedAt int64
}

type TTLStorage interface {
	SetTTL(key []byte, expiresAt int64) error
	GetTTL(key []byte) (*TTLMetadata, error)
	RemoveTTL(key []byte) error
	GetExpiredKeys(before int64) ([][]byte, error)
	RemoveTTLBatch(keys [][]byte) error
}

type LMDBTTLStorage struct {
	env    *lmdb.Env
	ttlDBI lmdb.DBI
}

func NewLMDBTTLStorage(env *lmdb.Env) (*LMDBTTLStorage, error) {
	var ttlDBI lmdb.DBI

	err := env.Update(func(txn *lmdb.Txn) error {
		var openErr error
		ttlDBI, openErr = txn.OpenDBI(TTLDatabaseName, lmdb.Create)
		return openErr
	})

	if HasError(err) {
		return nil, err
	}

	return &LMDBTTLStorage{
		env:    env,
		ttlDBI: ttlDBI,
	}, nil
}

func (ttlStorage *LMDBTTLStorage) SetTTL(key []byte, expiresAt int64) error {
	err := validateTTLKey(key)
	if HasError(err) {
		return err
	}

	err = validateTimestamp(expiresAt)
	if HasError(err) {
		return err
	}

	metadata := &TTLMetadata{
		Key:       key,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now().Unix(),
	}

	serialized := serializeTTLMetadata(metadata)

	return ttlStorage.env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(ttlStorage.ttlDBI, key, serialized, NoFlags)
	})
}

func (ttlStorage *LMDBTTLStorage) GetTTL(key []byte) (*TTLMetadata, error) {
	err := validateTTLKey(key)
	if HasError(err) {
		return nil, err
	}

	var metadata *TTLMetadata

	err = ttlStorage.env.View(func(txn *lmdb.Txn) error {
		value, getErr := txn.Get(ttlStorage.ttlDBI, key)

		if isNotFound(getErr) {
			return ErrTTLNotFound
		}

		if HasError(getErr) {
			return getErr
		}

		var deserializeErr error
		metadata, deserializeErr = deserializeTTLMetadata(value, key)
		return deserializeErr
	})

	if HasError(err) {
		return nil, err
	}

	return metadata, nil
}

func (ttlStorage *LMDBTTLStorage) RemoveTTL(key []byte) error {
	err := validateTTLKey(key)
	if HasError(err) {
		return err
	}

	return ttlStorage.env.Update(func(txn *lmdb.Txn) error {
		delErr := txn.Del(ttlStorage.ttlDBI, key, nil)

		if isNotFound(delErr) {
			return ErrTTLNotFound
		}

		return delErr
	})
}

func (ttlStorage *LMDBTTLStorage) GetExpiredKeys(before int64) ([][]byte, error) {
	err := validateTimestamp(before)
	if HasError(err) {
		return nil, err
	}

	var expiredKeys [][]byte

	err = ttlStorage.env.View(func(txn *lmdb.Txn) error {
		cursor, cursorErr := txn.OpenCursor(ttlStorage.ttlDBI)
		if HasError(cursorErr) {
			return cursorErr
		}
		defer cursor.Close()

		count := 0
		for {
			if count >= MaxCleanupBatchSize {
				break
			}

			key, value, cursorErr := cursor.Get(nil, nil, lmdb.Next)
			if isNotFound(cursorErr) {
				break
			}
			if HasError(cursorErr) {
				return cursorErr
			}

			metadata, deserializeErr := deserializeTTLMetadata(value, key)
			if HasError(deserializeErr) {
				continue
			}

			if isExpiredBefore(metadata.ExpiresAt, before) {
				keyCopy := make([]byte, len(key))
				copy(keyCopy, key)
				expiredKeys = append(expiredKeys, keyCopy)
				count++
			}
		}

		return nil
	})

	if HasError(err) {
		return nil, err
	}

	return expiredKeys, nil
}

func (ttlStorage *LMDBTTLStorage) RemoveTTLBatch(keys [][]byte) error {
	if IsEmpty(keys) {
		return nil
	}

	return ttlStorage.env.Update(func(txn *lmdb.Txn) error {
		for _, key := range keys {
			err := validateTTLKey(key)
			if HasError(err) {
				continue
			}

			delErr := txn.Del(ttlStorage.ttlDBI, key, nil)
			if isNotFound(delErr) {
				continue
			}
			if HasError(delErr) {
				return delErr
			}
		}
		return nil
	})
}
