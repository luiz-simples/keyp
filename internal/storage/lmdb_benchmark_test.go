package storage

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func BenchmarkTTLOperations(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "keyp_benchmark_ttl_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewLMDBStorage(filepath.Join(tempDir, "benchmark.db"))
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ttlManager := NewLMDBTTLManager(storage)

	b.Run("SetExpire", func(b *testing.B) {
		keys := make([][]byte, b.N)
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.SetExpire(keys[i], 3600)
		}
	})

	b.Run("SetExpireAt", func(b *testing.B) {
		keys := make([][]byte, b.N)
		timestamp := time.Now().Add(time.Hour).Unix()
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_expireat_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.SetExpireAt(keys[i], timestamp)
		}
	})

	b.Run("GetTTL", func(b *testing.B) {
		keys := make([][]byte, b.N)
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_ttl_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
			ttlManager.SetExpire(key, 3600)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.GetTTL(keys[i])
		}
	})

	b.Run("GetPTTL", func(b *testing.B) {
		keys := make([][]byte, b.N)
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_pttl_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
			ttlManager.SetExpire(key, 3600)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.GetPTTL(keys[i])
		}
	})

	b.Run("Persist", func(b *testing.B) {
		keys := make([][]byte, b.N)
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_persist_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
			ttlManager.SetExpire(key, 3600)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.Persist(keys[i])
		}
	})

	b.Run("IsExpired", func(b *testing.B) {
		keys := make([][]byte, b.N)
		for i := 0; i < b.N; i++ {
			key := []byte(fmt.Sprintf("benchmark_expired_key_%d", i))
			keys[i] = key
			storage.Set(key, []byte("benchmark_value"))
			ttlManager.SetExpire(key, 3600)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			ttlManager.IsExpired(keys[i])
		}
	})
}

func BenchmarkTTLCleanup(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "keyp_benchmark_cleanup_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewLMDBStorage(filepath.Join(tempDir, "cleanup.db"))
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ttlManager := NewLMDBTTLManager(storage)

	b.Run("CleanupExpired_Small", func(b *testing.B) {
		benchmarkCleanupWithSize(b, storage, ttlManager, 100)
	})

	b.Run("CleanupExpired_Medium", func(b *testing.B) {
		benchmarkCleanupWithSize(b, storage, ttlManager, 1000)
	})

	b.Run("CleanupExpired_Large", func(b *testing.B) {
		benchmarkCleanupWithSize(b, storage, ttlManager, 10000)
	})
}

func benchmarkCleanupWithSize(b *testing.B, storage *LMDBStorage, ttlManager *LMDBTTLManager, keyCount int) {
	for i := 0; i < b.N; i++ {
		b.StopTimer()

		for j := 0; j < keyCount; j++ {
			key := []byte(fmt.Sprintf("cleanup_key_%d_%d", i, j))
			storage.Set(key, []byte("cleanup_value"))
			ttlManager.SetExpire(key, 1)
		}

		time.Sleep(2 * time.Second)

		b.StartTimer()
		ttlManager.CleanupExpired()
	}
}

func BenchmarkTTLConcurrent(b *testing.B) {
	tempDir, err := os.MkdirTemp("", "keyp_benchmark_concurrent_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	storage, err := NewLMDBStorage(filepath.Join(tempDir, "concurrent.db"))
	if err != nil {
		b.Fatalf("Failed to create storage: %v", err)
	}
	defer storage.Close()

	ttlManager := NewLMDBTTLManager(storage)

	b.Run("ConcurrentSetExpire", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("concurrent_key_%d", i))
				storage.Set(key, []byte("concurrent_value"))
				ttlManager.SetExpire(key, 3600)
				i++
			}
		})
	})

	b.Run("ConcurrentGetTTL", func(b *testing.B) {
		for i := 0; i < 1000; i++ {
			key := []byte(fmt.Sprintf("concurrent_ttl_key_%d", i))
			storage.Set(key, []byte("concurrent_value"))
			ttlManager.SetExpire(key, 3600)
		}

		b.ResetTimer()
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := []byte(fmt.Sprintf("concurrent_ttl_key_%d", i%1000))
				ttlManager.GetTTL(key)
				i++
			}
		})
	})
}

func BenchmarkTTLMetrics(b *testing.B) {
	metrics := NewTTLMetrics()

	b.Run("RecordCleanupOperations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			startTime := metrics.RecordCleanupStart()
			metrics.RecordCleanupEnd(startTime, 10)
		}
	})

	b.Run("RecordTTLOperations", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics.RecordTTLSet()
			metrics.RecordTTLRemoved()
		}
	})

	b.Run("GetMetrics", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			metrics.GetCleanupOperations()
			metrics.GetKeysExpired()
			metrics.GetKeysWithTTL()
			metrics.GetAvgCleanupDuration()
		}
	})
}
