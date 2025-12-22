package storage

import (
	"sync/atomic"
	"time"
)

type TTLMetrics struct {
	cleanupOperations  int64
	cleanupDuration    int64
	keysExpired        int64
	keysWithTTL        int64
	cleanupErrors      int64
	lastCleanupTime    int64
	avgCleanupDuration int64
	totalCleanupTime   int64
}

func NewTTLMetrics() *TTLMetrics {
	return &TTLMetrics{}
}

func (metrics *TTLMetrics) RecordCleanupStart() int64 {
	return time.Now().UnixNano()
}

func (metrics *TTLMetrics) RecordCleanupEnd(startTime int64, keysExpired int) {
	duration := time.Now().UnixNano() - startTime

	atomic.AddInt64(&metrics.cleanupOperations, 1)
	atomic.AddInt64(&metrics.cleanupDuration, duration)
	atomic.AddInt64(&metrics.keysExpired, int64(keysExpired))
	atomic.StoreInt64(&metrics.lastCleanupTime, time.Now().Unix())

	totalTime := atomic.AddInt64(&metrics.totalCleanupTime, duration)
	operations := atomic.LoadInt64(&metrics.cleanupOperations)
	if operations > 0 {
		atomic.StoreInt64(&metrics.avgCleanupDuration, totalTime/operations)
	}
}

func (metrics *TTLMetrics) RecordCleanupError() {
	atomic.AddInt64(&metrics.cleanupErrors, 1)
}

func (metrics *TTLMetrics) RecordTTLSet() {
	atomic.AddInt64(&metrics.keysWithTTL, 1)
}

func (metrics *TTLMetrics) RecordTTLRemoved() {
	atomic.AddInt64(&metrics.keysWithTTL, -1)
}

func (metrics *TTLMetrics) GetCleanupOperations() int64 {
	return atomic.LoadInt64(&metrics.cleanupOperations)
}

func (metrics *TTLMetrics) GetKeysExpired() int64 {
	return atomic.LoadInt64(&metrics.keysExpired)
}

func (metrics *TTLMetrics) GetKeysWithTTL() int64 {
	return atomic.LoadInt64(&metrics.keysWithTTL)
}

func (metrics *TTLMetrics) GetCleanupErrors() int64 {
	return atomic.LoadInt64(&metrics.cleanupErrors)
}

func (metrics *TTLMetrics) GetLastCleanupTime() int64 {
	return atomic.LoadInt64(&metrics.lastCleanupTime)
}

func (metrics *TTLMetrics) GetAvgCleanupDuration() time.Duration {
	nanos := atomic.LoadInt64(&metrics.avgCleanupDuration)
	return time.Duration(nanos)
}

func (metrics *TTLMetrics) GetTotalCleanupTime() time.Duration {
	nanos := atomic.LoadInt64(&metrics.totalCleanupTime)
	return time.Duration(nanos)
}

func (metrics *TTLMetrics) Reset() {
	atomic.StoreInt64(&metrics.cleanupOperations, 0)
	atomic.StoreInt64(&metrics.cleanupDuration, 0)
	atomic.StoreInt64(&metrics.keysExpired, 0)
	atomic.StoreInt64(&metrics.keysWithTTL, 0)
	atomic.StoreInt64(&metrics.cleanupErrors, 0)
	atomic.StoreInt64(&metrics.lastCleanupTime, 0)
	atomic.StoreInt64(&metrics.avgCleanupDuration, 0)
	atomic.StoreInt64(&metrics.totalCleanupTime, 0)
}
