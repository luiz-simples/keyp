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
			invalidPath := "/nonexistent/path/that/should/not/exist"
			_, err := storage.NewLMDBStorage(invalidPath)
			Expect(err).To(HaveOccurred())
		})

		It("should handle storage creation with read-only directory", func() {
			readOnlyDir, err := os.MkdirTemp("", "readonly-test-*")
			Expect(err).NotTo(HaveOccurred())
			defer os.RemoveAll(readOnlyDir)

			err = os.Chmod(readOnlyDir, 0444)
			Expect(err).NotTo(HaveOccurred())

			_, err = storage.NewLMDBStorage(readOnlyDir)
			Expect(err).To(HaveOccurred())

			os.Chmod(readOnlyDir, 0755)
		})
	})

	Describe("LMDB Storage operation edge cases", func() {
		It("should handle Get operation edge cases", func() {
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			largeKey := make([]byte, storage.MaxKeySize+1)
			_, err = lmdbStorage.Get(largeKey)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key too large"))
		})

		It("should handle Set operation edge cases", func() {
			err := lmdbStorage.Set([]byte{}, []byte("value"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			largeKey := make([]byte, storage.MaxKeySize+1)
			err = lmdbStorage.Set(largeKey, []byte("value"))
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("key too large"))
		})

		It("should handle Del operation with no keys", func() {
			count, err := lmdbStorage.Del()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})

		It("should handle Del operation with empty keys", func() {
			keys := [][]byte{
				[]byte("valid_key"),
				[]byte{}, // empty key
				[]byte("another_valid_key"),
			}

			err := lmdbStorage.Set([]byte("valid_key"), []byte("value1"))
			Expect(err).NotTo(HaveOccurred())
			err = lmdbStorage.Set([]byte("another_valid_key"), []byte("value2"))
			Expect(err).NotTo(HaveOccurred())

			count, err := lmdbStorage.Del(keys...)
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(2)) // Only valid keys should be deleted
		})

		It("should handle Get operation with invalid keys", func() {
			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

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

			err := ttlStorage.SetTTL(testKey, -1)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("invalid timestamp"))

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
			err := ttlStorage.RemoveTTL([]byte("nonexistent"))
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(storage.ErrTTLNotFound))
		})

		It("should handle GetExpiredKeys with various scenarios", func() {
			baseTime := time.Now().Unix()

			keys := [][]byte{
				[]byte("expired:1"),
				[]byte("expired:2"),
				[]byte("future:1"),
				[]byte("future:2"),
			}

			err := ttlStorage.SetTTL(keys[0], baseTime-100)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[1], baseTime-50)
			Expect(err).NotTo(HaveOccurred())

			err = ttlStorage.SetTTL(keys[2], baseTime+100)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[3], baseTime+200)
			Expect(err).NotTo(HaveOccurred())

			expiredKeys, err := ttlStorage.GetExpiredKeys(baseTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(Equal(2))

			expiredKeyStrings := make([]string, len(expiredKeys))
			for i, key := range expiredKeys {
				expiredKeyStrings[i] = string(key)
			}
			Expect(expiredKeyStrings).To(ContainElement("expired:1"))
			Expect(expiredKeyStrings).To(ContainElement("expired:2"))
		})

		It("should handle RemoveTTLBatch with mixed scenarios", func() {
			keys := [][]byte{
				[]byte("batch:1"),
				[]byte("batch:2"),
				[]byte("batch:3"),
				[]byte("nonexistent"),
			}

			baseTime := time.Now().Unix()

			for i := range 3 {
				err := ttlStorage.SetTTL(keys[i], baseTime+int64(i*100))
				Expect(err).NotTo(HaveOccurred())
			}

			err := ttlStorage.RemoveTTLBatch(keys)
			Expect(err).NotTo(HaveOccurred())

			for i := range 3 {
				_, err := ttlStorage.GetTTL(keys[i])
				Expect(err).To(Equal(storage.ErrTTLNotFound))
			}
		})

		It("should handle empty batch operations", func() {
			err := ttlStorage.RemoveTTLBatch([][]byte{})
			Expect(err).NotTo(HaveOccurred())

			expiredKeys, err := ttlStorage.GetExpiredKeys(time.Now().Unix() - 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(Equal(0))
		})

		It("should handle GetExpiredKeys with no expired keys", func() {
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
			testKey := []byte("restore:test")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			ttlStorage := lmdbStorage.GetTTLStorage()
			pastTime := time.Now().Unix() - 100
			err = ttlStorage.SetTTL(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())

			err = ttlManager.RestoreTTL()
			Expect(err).NotTo(HaveOccurred())

			_, err = lmdbStorage.Get(testKey)
			Expect(err).To(Equal(storage.ErrKeyNotFound))
		})

		It("should handle CleanupExpired with no expired keys", func() {
			err := ttlManager.CleanupExpired()
			Expect(err).NotTo(HaveOccurred())
		})

		It("should handle SetExpire with non-existent key", func() {
			result, err := ttlManager.SetExpire([]byte("nonexistent"), 3600)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))
		})

		It("should handle SetExpireAt with edge cases", func() {
			testKey := []byte("expireat:test")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			currentTime := time.Now().Unix()
			result, err := ttlManager.SetExpireAt(testKey, currentTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))

			futureTime := currentTime + 1
			result, err = ttlManager.SetExpireAt(testKey, futureTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))
		})

		It("should handle GetTTL with various key states", func() {
			testKey := []byte("ttl:test")
			testValue := []byte("test:value")

			result, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLNotFound)))

			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLPersistent)))

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

			result, err := ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLNotFound)))

			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(int64(storage.TTLPersistent)))

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

			result, err := ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistFailure))

			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistFailure))

			_, err = ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())

			result, err = ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistSuccess))

			ttlResult, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttlResult).To(Equal(int64(storage.TTLPersistent)))
		})

		It("should handle IsExpired with various scenarios", func() {
			testKey := []byte("expired:check")
			testValue := []byte("test:value")

			expired, err := ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeFalse())

			err = lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			expired, err = ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeFalse())

			ttlStorage := lmdbStorage.GetTTLStorage()
			pastTime := time.Now().Unix() - 100
			err = ttlStorage.SetTTL(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())

			expired, err = ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeTrue())
		})

		It("should handle CleanupExpired with no expired keys", func() {
			err := ttlManager.CleanupExpired()
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("Utils edge cases", func() {
		It("should handle isEmpty with different data types", func() {

			_, err := lmdbStorage.Get([]byte{})
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("empty key"))

			count, err := lmdbStorage.Del()
			Expect(err).NotTo(HaveOccurred())
			Expect(count).To(Equal(0))
		})

		It("should handle deserializeTTLMetadata with corrupted data", func() {
			testKey := []byte("corrupt:test")

			ttlStorage := lmdbStorage.GetTTLStorage()
			err := ttlStorage.SetTTL(testKey, time.Now().Unix()+3600)
			Expect(err).NotTo(HaveOccurred())

			_, err = ttlStorage.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
