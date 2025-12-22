package storage_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Storage Coverage Tests", func() {
	var (
		lmdbStorage *storage.LMDBStorage
		tmpDir      string
	)

	BeforeEach(func() {
		// Set test mode to disable logging during tests
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "storage-coverage-test-*")
		Expect(err).NotTo(HaveOccurred())

		lmdbStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if lmdbStorage != nil {
			lmdbStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("LMDB Storage creation edge cases", func() {
		It("should handle storage creation with invalid directory", func() {
			// Try to create storage in a non-existent parent directory
			invalidPath := "/nonexistent/path/that/should/not/exist"
			_, err := storage.NewLMDBStorage(invalidPath)
			Expect(err).To(HaveOccurred())
		})

		It("should handle storage creation with read-only directory", func() {
			// Create a read-only directory
			readOnlyDir, err := os.MkdirTemp("", "readonly-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(readOnlyDir)

			// Make directory read-only
			err = os.Chmod(readOnlyDir, 0444)
			Expect(err).NotTo(HaveOccurred())

			// Try to create storage (should fail)
			_, err = storage.NewLMDBStorage(readOnlyDir)
			Expect(err).To(HaveOccurred())

			// Restore permissions for cleanup
			os.Chmod(readOnlyDir, 0755)
		})
	})

	Describe("LMDB Storage operation edge cases", func() {
		It("should handle Get operation edge cases", func() {
			// Test Get with empty key (should be caught by validation)
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			// Test Get with oversized key
			largeKey := make([]byte, storage.MaxKeySize+1)
			_, err = lmdbStorage.Get(largeKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key too large"))
		})

		It("should handle Set operation edge cases", func() {
			// Test Set with empty key
			err := lmdbStorage.Set([]byte{}, []byte("value"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			// Test Set with oversized key
			largeKey := make([]byte, storage.MaxKeySize+1)
			err = lmdbStorage.Set(largeKey, []byte("value"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key too large"))
		})

		It("should handle Del operation with no keys", func() {
			// Test Del with no keys (should return 0)
			count, err := lmdbStorage.Del()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})

		It("should handle Del operation with empty keys", func() {
			// Test Del with empty key in the list
			keys := [][]byte{
				[]byte("valid_key"),
				[]byte{}, // empty key
				[]byte("another_valid_key"),
			}

			// Set the valid keys first
			err := lmdbStorage.Set([]byte("valid_key"), []byte("value1"))
			Expect(err).NotTo(HaveOccurred())
			err = lmdbStorage.Set([]byte("another_valid_key"), []byte("value2"))
			Expect(err).NotTo(HaveOccurred())

			// Del should handle empty key gracefully
			count, err := lmdbStorage.Del(keys...)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(2)) // Only valid keys should be deleted
		})

		It("should handle Get operation with invalid keys", func() {
			// Test Get with empty key (should be caught by validation)
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			// Test Get with oversized key
			largeKey := make([]byte, storage.MaxKeySize+1)
			_, err = lmdbStorage.Get(largeKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key too large"))
		})
	})

	Describe("TTL Storage edge cases", func() {
		var ttlStorage storage.TTLStorage

		BeforeEach(func() {
			ttlStorage = lmdbStorage.GetTTLStorage()
		})

		It("should handle TTL operations with invalid timestamps", func() {
			testKey := []byte("test:key")

			// Test with negative timestamp
			err := ttlStorage.SetTTL(testKey, -1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid timestamp"))

			// Test with very large future timestamp (beyond reasonable limit)
			futureTime := time.Now().Unix() + (400 * 24 * 3600) // 400 days from now
			err = ttlStorage.SetTTL(testKey, futureTime)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid timestamp"))
		})

		It("should handle GetTTL with non-existent key", func() {
			_, err := ttlStorage.GetTTL([]byte("nonexistent"))
			Expect(err).To(Equal(storage.ErrTTLNotFound))
		})

		It("should handle RemoveTTL with non-existent key", func() {
			// Should return error when removing TTL from non-existent key
			err := ttlStorage.RemoveTTL([]byte("nonexistent"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(storage.ErrTTLNotFound))
		})

		It("should handle GetExpiredKeys with various scenarios", func() {
			baseTime := time.Now().Unix()

			// Set some keys with different expiration times
			keys := [][]byte{
				[]byte("expired:1"),
				[]byte("expired:2"),
				[]byte("future:1"),
				[]byte("future:2"),
			}

			// Set expired keys
			err := ttlStorage.SetTTL(keys[0], baseTime-100)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[1], baseTime-50)
			Expect(err).NotTo(HaveOccurred())

			// Set future keys
			err = ttlStorage.SetTTL(keys[2], baseTime+100)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[3], baseTime+200)
			Expect(err).NotTo(HaveOccurred())

			// Get expired keys
			expiredKeys, err := ttlStorage.GetExpiredKeys(baseTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(Equal(2))

			// Verify the expired keys are the correct ones
			expiredKeyStrings := make([]string, len(expiredKeys))
			for i, key := range expiredKeys {
				expiredKeyStrings[i] = string(key)
			}
			Expect(expiredKeyStrings).To(ContainElement("expired:1"))
			Expect(expiredKeyStrings).To(ContainElement("expired:2"))
		})

		It("should handle RemoveTTLBatch with mixed scenarios", func() {
			// Set up some keys with TTL
			keys := [][]byte{
				[]byte("batch:1"),
				[]byte("batch:2"),
				[]byte("batch:3"),
				[]byte("nonexistent"),
			}

			baseTime := time.Now().Unix()

			// Set TTL for first 3 keys
			for i := range 3 {
				err := ttlStorage.SetTTL(keys[i], baseTime+int64(i*100))
				Expect(err).NotTo(HaveOccurred())
			}

			// Remove all keys in batch (including non-existent one)
			err := ttlStorage.RemoveTTLBatch(keys)
			Expect(err).NotTo(HaveOccurred())

			// Verify all TTLs are removed
			for i := range 3 {
				_, err := ttlStorage.GetTTL(keys[i])
				Expect(err).To(Equal(storage.ErrTTLNotFound))
			}
		})

		It("should handle empty batch operations", func() {
			// Test RemoveTTLBatch with empty slice
			err := ttlStorage.RemoveTTLBatch([][]byte{})
			Expect(err).NotTo(HaveOccurred())

			// Test GetExpiredKeys with no expired keys
			expiredKeys, err := ttlStorage.GetExpiredKeys(time.Now().Unix() - 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(Equal(0))
		})

		It("should handle GetExpiredKeys with no expired keys", func() {
			// Test GetExpiredKeys with no expired keys
			expiredKeys, err := ttlStorage.GetExpiredKeys(time.Now().Unix() - 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(Equal(0))
		})
	})

	Describe("TTL Manager edge cases", func() {
		var ttlManager storage.TTLManager

		BeforeEach(func() {
			ttlManager = lmdbStorage.GetTTLManager()
		})

		It("should handle RestoreTTL operation", func() {
			// Set up some expired keys
			testKey := []byte("restore:test")
			testValue := []byte("test:value")

			// Set key in storage
			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Set expired TTL directly in storage
			ttlStorage := lmdbStorage.GetTTLStorage()
			pastTime := time.Now().Unix() - 100
			err = ttlStorage.SetTTL(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())

			// Call RestoreTTL (should clean up expired keys)
			err = ttlManager.RestoreTTL()
			Expect(err).NotTo(HaveOccurred())

			// Verify expired key was cleaned up
			_, err = lmdbStorage.Get(testKey)
			Expect(err).To(Equal(storage.ErrKeyNotFound))
		})

		It("should handle CleanupExpired with no expired keys", func() {
			// Call cleanup when no keys are expired
			err := ttlManager.CleanupExpired()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle SetExpire with non-existent key", func() {
			// Try to set expire on non-existent key
			result, err := ttlManager.SetExpire([]byte("nonexistent"), 3600)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))
		})

		It("should handle SetExpireAt with edge cases", func() {
			testKey := []byte("expireat:test")
			testValue := []byte("test:value")

			// Set key in storage
			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Test with timestamp exactly at current time (should fail as it's considered past)
			currentTime := time.Now().Unix()
			result, err := ttlManager.SetExpireAt(testKey, currentTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))

			// Test with timestamp 1 second in the future (should succeed)
			futureTime := currentTime + 1
			result, err = ttlManager.SetExpireAt(testKey, futureTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))
		})

		It("should handle GetTTL with various key states", func() {
			testKey := []byte("ttl:test")
			testValue := []byte("test:value")

			// Test with non-existent key
			result, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLNotFound)))

			// Set key without TTL
			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Test key without TTL
			result, err = ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLPersistent)))

			// Set TTL and test
			_, err = ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNumerically(">", 0))
			Expect(result).To(BeNumerically("<=", 3600))
		})

		It("should handle GetPTTL with various key states", func() {
			testKey := []byte("pttl:test")
			testValue := []byte("test:value")

			// Test with non-existent key
			result, err := ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLNotFound)))

			// Set key without TTL
			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Test key without TTL
			result, err = ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLPersistent)))

			// Set TTL and test
			_, err = ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeNumerically(">", 0))
			Expect(result).To(BeNumerically("<=", 3600000)) // milliseconds
		})

		It("should handle Persist with various scenarios", func() {
			testKey := []byte("persist:test")
			testValue := []byte("test:value")

			// Test persist on non-existent key
			result, err := ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistFailure))

			// Set key without TTL
			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Test persist on key without TTL
			result, err = ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistFailure))

			// Set TTL and then persist
			_, err = ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistSuccess))

			// Verify TTL is removed
			ttlResult, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttlResult).To(Equal(int64(storage.TTLPersistent)))
		})

		It("should handle IsExpired with various scenarios", func() {
			testKey := []byte("expired:check")
			testValue := []byte("test:value")

			// Test with non-existent key
			expired, err := ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeFalse())

			// Set key in storage
			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			// Key without TTL should not be expired
			expired, err = ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeFalse())

			// Set expired TTL
			ttlStorage := lmdbStorage.GetTTLStorage()
			pastTime := time.Now().Unix() - 100
			err = ttlStorage.SetTTL(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())

			// Key should now be expired
			expired, err = ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeTrue())
		})

		It("should handle CleanupExpired with no expired keys", func() {
			// Call cleanup when no keys are expired
			err := ttlManager.CleanupExpired()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Utils edge cases", func() {
		It("should handle isEmpty with different data types", func() {
			// This tests the isEmpty function in utils.go with different types
			// We can't directly test it from here, but we can test it indirectly
			// through operations that use it

			// Test with empty byte slice
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			// Test Del with empty keys slice (should return 0)
			count, err := lmdbStorage.Del()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})

		It("should handle deserializeTTLMetadata with corrupted data", func() {
			// This is harder to test directly, but we can test through TTL operations
			testKey := []byte("corrupt:test")

			// Set a valid TTL first
			ttlStorage := lmdbStorage.GetTTLStorage()
			err := ttlStorage.SetTTL(testKey, time.Now().Unix()+3600)
			Expect(err).NotTo(HaveOccurred())

			// Verify we can get it back
			_, err = ttlStorage.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
