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

const MaxKeySize = 511

type LMDBStorage struct {
	env *lmdb.Env
	dbi lmdb.DBI
}

func NewLMDBStorage(dataDir string) (*LMDBStorage, error) {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return nil, err
	}

	env, err := lmdb.NewEnv()
	if err != nil {
		return nil, err
	}

	err = env.SetMaxDBs(1)
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	err = env.SetMapSize(1 << 30)
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	err = env.Open(dataDir, 0, 0644)
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	var dbi lmdb.DBI
	err = env.Update(func(txn *lmdb.Txn) error {
		dbi, err = txn.OpenDBI("keyp", lmdb.Create)
		return err
	})
	if err != nil {
		_ = env.Close()
		return nil, err
	}

	return &LMDBStorage{
		env: env,
		dbi: dbi,
	}, nil
}

func (storage *LMDBStorage) Close() error {
	return storage.env.Close()
}

func isEmpty(key []byte) bool {
	return len(key) == 0
}

func exceedsLimit(key []byte) bool {
	return len(key) > MaxKeySize
}

func (storage *LMDBStorage) validateKey(key []byte) error {
	if isEmpty(key) {
		return ErrEmptyKey
	}

	if exceedsLimit(key) {
		return ErrKeyTooLarge
	}

	return nil
}

func (storage *LMDBStorage) Set(key, value []byte) error {
	err := storage.validateKey(key)
	if err != nil {
		return err
	}

	return storage.env.Update(func(txn *lmdb.Txn) error {
		return txn.Put(storage.dbi, key, value, 0)
	})
}

func (storage *LMDBStorage) Get(key []byte) ([]byte, error) {
	err := storage.validateKey(key)
	if err != nil {
		return nil, err
	}

	var value []byte

	err = storage.env.View(func(txn *lmdb.Txn) error {
		val, getErr := txn.Get(storage.dbi, key)

		if lmdb.IsNotFound(getErr) {
			return ErrKeyNotFound
		}

		if getErr != nil {
			return getErr
		}

		value = make([]byte, len(val))
		copy(value, val)

		return nil
	})

	if err != nil {
		return nil, err
	}

	return value, nil
}

func (storage *LMDBStorage) Del(keys ...[]byte) (int, error) {
	deleted := 0

	err := storage.env.Update(func(txn *lmdb.Txn) error {
		for _, key := range keys {
			err := storage.validateKey(key)

			if err != nil {
				continue
			}

			err = txn.Del(storage.dbi, key, nil)

			if lmdb.IsNotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			deleted++
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return deleted, nil
}
