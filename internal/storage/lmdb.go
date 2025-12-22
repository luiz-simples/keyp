package storage

import (
	"errors"
	"os"

	"github.com/PowerDNS/lmdb-go/lmdb"
)

var (
	ErrKeyNotFound = errors.New("key not found")
	ErrEmptyKey    = errors.New("empty key")
	ErrKeyTooLarge = errors.New("key too large")
)

const (
	MaxKeySize         = 511
	DirPermissions     = 0755
	FilePermissions    = 0644
	MaxDatabases       = 1
	MapSizeBytes       = 1 << 30
	NoFlags            = 0
	InitialDeleteCount = 0
)

type LMDBStorage struct {
	env        *lmdb.Env
	dbi        lmdb.DBI
	ttlStorage *LMDBTTLStorage
	ttlManager *LMDBTTLManager
}

func NewLMDBStorage(dataDir string) (*LMDBStorage, error) {
	err := os.MkdirAll(dataDir, DirPermissions)
	if HasError(err) {
		return nil, err
	}

	env, err := lmdb.NewEnv()
	if HasError(err) {
		return nil, err
	}

	err = env.SetMaxDBs(MaxDatabases + 1)
	if HasError(err) {
		env.Close()
		return nil, err
	}

	err = env.SetMapSize(MapSizeBytes)
	if HasError(err) {
		env.Close()
		return nil, err
	}

	err = env.Open(dataDir, NoFlags, FilePermissions)
	if HasError(err) {
		env.Close()
		return nil, err
	}

	var dbi lmdb.DBI
	err = env.Update(func(txn *lmdb.Txn) error {
		dbi, err = txn.OpenDBI("keyp", lmdb.Create)
		return err
	})
	if HasError(err) {
		env.Close()
		return nil, err
	}

	ttlStorage, err := NewLMDBTTLStorage(env)
	if HasError(err) {
		env.Close()
		return nil, err
	}

	storage := &LMDBStorage{
		env:        env,
		dbi:        dbi,
		ttlStorage: ttlStorage,
	}

	ttlManager := NewLMDBTTLManager(storage)
	storage.ttlManager = ttlManager

	return storage, nil
}

func (storage *LMDBStorage) Close() error {
	return storage.env.Close()
}

func (storage *LMDBStorage) validateKey(key []byte) error {
	if IsEmpty(key) {
		return ErrEmptyKey
	}

	if exceedsLimit(key) {
		return ErrKeyTooLarge
	}

	return nil
}

func (storage *LMDBStorage) Set(key, value []byte) error {
	err := storage.validateKey(key)
	if HasError(err) {
		return err
	}

	return storage.env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(storage.dbi, key, value, NoFlags)
	})
}

func (storage *LMDBStorage) Get(key []byte) ([]byte, error) {
	err := storage.validateKey(key)
	if HasError(err) {
		return nil, err
	}

	var value []byte

	err = storage.env.View(func(txn *lmdb.Txn) error {
		ttlValue, ttlErr := txn.Get(storage.ttlStorage.GetTTLDBI(), key)

		if !isNotFound(ttlErr) && !HasError(ttlErr) {
			if isExpiredFromTTLData(ttlValue) {
				return ErrKeyNotFound
			}
		}

		val, getErr := txn.Get(storage.dbi, key)

		if isNotFound(getErr) {
			return ErrKeyNotFound
		}

		if HasError(getErr) {
			return getErr
		}

		value = val
		return nil
	})

	if HasError(err) {
		if err == ErrKeyNotFound {
			storage.cleanupExpiredKey(key)
		}
		return nil, err
	}

	return value, nil
}

func (storage *LMDBStorage) Del(keys ...[]byte) (int, error) {
	deleted := InitialDeleteCount

	err := storage.env.Update(func(txn *lmdb.Txn) error {
		for _, key := range keys {
			err := storage.validateKey(key)

			if HasError(err) {
				continue
			}

			err = txn.Del(storage.dbi, key, nil)

			if isNotFound(err) {
				continue
			}

			if HasError(err) {
				return err
			}

			deleted++
		}

		return nil
	})

	if HasError(err) {
		return InitialDeleteCount, err
	}

	for _, key := range keys {
		storage.cleanupTTLMetadata(key)
	}

	return deleted, nil
}

func (storage *LMDBStorage) GetTTLStorage() TTLStorage {
	return storage.ttlStorage
}

func (storage *LMDBStorage) GetTTLManager() TTLManager {
	return storage.ttlManager
}

func (storage *LMDBStorage) cleanupExpiredKey(key []byte) {
	_, err := storage.Del(key)
	if HasError(err) {
		return
	}
}

func (storage *LMDBStorage) cleanupTTLMetadata(key []byte) {
	err := storage.ttlStorage.RemoveTTL(key)
	if HasError(err) {
		return
	}
}
