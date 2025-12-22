package storage

import (
	"time"
)

const (
	TTLNotFound         = -2
	TTLPersistent       = -1
	ExpireSuccess       = 1
	ExpireFailure       = 0
	PersistSuccess      = 1
	PersistFailure      = 0
	MaxCleanupBatchSize = 1000
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
	metrics    *TTLMetrics
}

func NewLMDBTTLManager(storage *LMDBStorage) *LMDBTTLManager {
	return &LMDBTTLManager{
		storage:    storage,
		ttlStorage: storage.GetTTLStorage(),
		metrics:    NewTTLMetrics(),
	}
}

func (manager *LMDBTTLManager) SetExpire(key []byte, seconds int64) (int, error) {
	err := validateTTLKey(key)
	if HasError(err) {
		return ExpireFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return ExpireFailure, nil
	}
	if HasError(err) {
		return ExpireFailure, err
	}

	if isNegativeSeconds(seconds) {
		return ExpireFailure, nil
	}

	expiresAt := calculateExpiresAt(seconds)

	err = manager.ttlStorage.SetTTL(key, expiresAt)
	if HasError(err) {
		return ExpireFailure, err
	}

	manager.metrics.RecordTTLSet()
	return ExpireSuccess, nil
}

func (manager *LMDBTTLManager) SetExpireAt(key []byte, timestamp int64) (int, error) {
	err := validateTTLKey(key)
	if HasError(err) {
		return ExpireFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return ExpireFailure, nil
	}
	if HasError(err) {
		return ExpireFailure, err
	}

	if isPastTimestamp(timestamp) {
		return ExpireFailure, nil
	}

	err = manager.ttlStorage.SetTTL(key, timestamp)
	if HasError(err) {
		return ExpireFailure, err
	}

	manager.metrics.RecordTTLSet()
	return ExpireSuccess, nil
}

func (manager *LMDBTTLManager) GetTTL(key []byte) (int64, error) {
	err := validateTTLKey(key)
	if HasError(err) {
		return TTLNotFound, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return TTLNotFound, nil
	}
	if HasError(err) {
		return TTLNotFound, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return TTLPersistent, nil
	}
	if HasError(err) {
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
	if HasError(err) {
		return TTLNotFound, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return TTLNotFound, nil
	}
	if HasError(err) {
		return TTLNotFound, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return TTLPersistent, nil
	}
	if HasError(err) {
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
	if HasError(err) {
		return PersistFailure, err
	}

	_, err = manager.storage.Get(key)
	if isNotFound(err) {
		return PersistFailure, nil
	}
	if HasError(err) {
		return PersistFailure, err
	}

	err = manager.ttlStorage.RemoveTTL(key)
	if err == ErrTTLNotFound {
		return PersistFailure, nil
	}
	if HasError(err) {
		return PersistFailure, err
	}

	manager.metrics.RecordTTLRemoved()
	return PersistSuccess, nil
}

func (manager *LMDBTTLManager) IsExpired(key []byte) (bool, error) {
	err := validateTTLKey(key)
	if HasError(err) {
		return false, err
	}

	metadata, err := manager.ttlStorage.GetTTL(key)
	if err == ErrTTLNotFound {
		return false, nil
	}
	if HasError(err) {
		return false, err
	}

	return isKeyExpired(metadata.ExpiresAt), nil
}

func (manager *LMDBTTLManager) CleanupExpired() error {
	startTime := manager.metrics.RecordCleanupStart()

	now := time.Now().Unix()

	expiredKeys, err := manager.ttlStorage.GetExpiredKeys(now)
	if HasError(err) {
		manager.metrics.RecordCleanupError()
		return err
	}

	if IsEmpty(expiredKeys) {
		manager.metrics.RecordCleanupEnd(startTime, 0)
		return nil
	}

	keysExpired := len(expiredKeys)

	_, err = manager.storage.Del(expiredKeys...)
	if HasError(err) {
		manager.metrics.RecordCleanupError()
		return err
	}

	err = manager.ttlStorage.RemoveTTLBatch(expiredKeys)
	if HasError(err) {
		manager.metrics.RecordCleanupError()
		return err
	}

	manager.metrics.RecordCleanupEnd(startTime, keysExpired)
	return nil
}

func (manager *LMDBTTLManager) RestoreTTL() error {
	return manager.CleanupExpired()
}

func (manager *LMDBTTLManager) GetMetrics() *TTLMetrics {
	return manager.metrics
}
