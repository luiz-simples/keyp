package storage

import (
	"time"
)

const (
	TTLNotFound    = -2
	TTLPersistent  = -1
	ExpireSuccess  = 1
	ExpireFailure  = 0
	PersistSuccess = 1
	PersistFailure = 0
)

type TTLManager interface {
	SetExpire(key []byte, seconds int64) (int, error)
	SetExpireAt(key []byte, timestamp int64) (int, error)
	GetTTL(key []byte) (int64, error)
	GetPTTL(key []byte) (int64, error)
	Persist(key []byte) (int, error)
	IsExpired(key []byte) (bool, error)
	CleanupExpired() error
	RestoreTTL() error
}

type LMDBTTLManager struct {
	storage    *LMDBStorage
	ttlStorage TTLStorage
}

func NewLMDBTTLManager(storage *LMDBStorage) *LMDBTTLManager {
	return &LMDBTTLManager{
		storage:    storage,
		ttlStorage: storage.GetTTLStorage(),
	}
}

func (manager *LMDBTTLManager) SetExpire(key []byte, seconds int64) (int, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return ExpireFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return ExpireFailure, nil
	}
	if hasError(err) {
		return ExpireFailure, err
	}

	if isNegativeSeconds(seconds) {
		return ExpireFailure, nil
	}

	expiresAt := calculateExpiresAt(seconds)

	err = manager.ttlStorage.SetTTL(key, expiresAt)
	if hasError(err) {
		return ExpireFailure, err
	}

	return ExpireSuccess, nil
}

func (manager *LMDBTTLManager) SetExpireAt(key []byte, timestamp int64) (int, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return ExpireFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return ExpireFailure, nil
	}
	if hasError(err) {
		return ExpireFailure, err
	}

	if isPastTimestamp(timestamp) {
		return ExpireFailure, nil
	}

	err = manager.ttlStorage.SetTTL(key, timestamp)
	if hasError(err) {
		return ExpireFailure, err
	}

	return ExpireSuccess, nil
}

func (manager *LMDBTTLManager) GetTTL(key []byte) (int64, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return TTLNotFound, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return TTLNotFound, nil
	}
	if hasError(err) {
		return TTLNotFound, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return TTLPersistent, nil
	}
	if hasError(err) {
		return TTLNotFound, err
	}

	remainingSeconds := calculateRemainingSeconds(metadata.ExpiresAt)
	if isExpiredTime(remainingSeconds) {
		return TTLNotFound, nil
	}

	return remainingSeconds, nil
}

func (manager *LMDBTTLManager) GetPTTL(key []byte) (int64, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return TTLNotFound, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return TTLNotFound, nil
	}
	if hasError(err) {
		return TTLNotFound, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return TTLPersistent, nil
	}
	if hasError(err) {
		return TTLNotFound, err
	}

	remainingMilliseconds := calculateRemainingMilliseconds(metadata.ExpiresAt)
	if isExpiredTime(remainingMilliseconds) {
		return TTLNotFound, nil
	}

	return remainingMilliseconds, nil
}

func (manager *LMDBTTLManager) Persist(key []byte) (int, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return PersistFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return PersistFailure, nil
	}
	if hasError(err) {
		return PersistFailure, err
	}

	err = manager.ttlStorage.RemoveTTL(key)
	if err == ErrTTLNotFound {
		return PersistFailure, nil
	}
	if hasError(err) {
		return PersistFailure, err
	}

	return PersistSuccess, nil
}

func (manager *LMDBTTLManager) IsExpired(key []byte) (bool, error) {
	err := validateTTLKey(key)
	if hasError(err) {
		return false, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return false, nil
	}
	if hasError(err) {
		return false, err
	}

	return isKeyExpired(metadata.ExpiresAt), nil
}

func (manager *LMDBTTLManager) CleanupExpired() error {
	now := time.Now().Unix()

	expiredKeys, err := manager.ttlStorage.GetExpiredKeys(now)
	if hasError(err) {
		return err
	}

	if isEmpty(expiredKeys) {
		return nil
	}

	_, err = manager.storage.Del(expiredKeys...)
	if hasError(err) {
		return err
	}

	err = manager.ttlStorage.RemoveTTLBatch(expiredKeys)
	if hasError(err) {
		return err
	}

	return nil
}

func (manager *LMDBTTLManager) RestoreTTL() error {
	return manager.CleanupExpired()
}
