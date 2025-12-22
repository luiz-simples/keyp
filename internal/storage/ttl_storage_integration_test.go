package storage_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("TTL Storage Integration Tests", func() {
	var (
		testStorage *storage.LMDBStorage
		ttlManager  storage.TTLManager
		tmpDir      string
	)

	BeforeEach(func() {
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "ttl-storage-integration-*")
		Expect(err).NotTo(HaveOccurred())

		testStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		ttlManager = testStorage.GetTTLManager()
	})

	AfterEach(func() {
		if testStorage != nil {
			testStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("TTL persistence across restarts", func() {
		It("should persist TTL metadata across storage restarts", func() {
			key := []byte("persistent_ttl_key")
			value := []byte("persistent_value")
			ttlSeconds := int64(3600)

			err := testStorage.Set(key, value)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.SetExpire(key, ttlSeconds)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))

			ttl, err := ttlManager.GetTTL(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically(">", 3500))

			originalTmpDir := tmpDir
			testStorage.Close()

			newStorage, err := storage.NewLMDBStorage(originalTmpDir)
			Expect(err).NotTo(HaveOccurred())
			defer newStorage.Close()

			newTTLManager := newStorage.GetTTLManager()

			retrievedValue, err := newStorage.Get(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))

			newTTL, err := newTTLManager.GetTTL(key)
			Expect(err).NotTo(HaveOccurred())
			Expect(newTTL).To(BeNumerically(">", 3400))
			Expect(newTTL).To(BeNumerically("<=", ttl))
		})

		It("should cleanup expired keys during startup", func() {
			expiredKey := []byte("expired_startup_key")
			activeKey := []byte("active_startup_key")
			value := []byte("startup_value")

			err := testStorage.Set(expiredKey, value)
			Expect(err).NotTo(HaveOccurred())
			err = testStorage.Set(activeKey, value)
			Expect(err).NotTo(HaveOccurred())

			pastTime := time.Now().Unix() - 100
			futureTime := time.Now().Unix() + 3600

			ttlStorage := testStorage.GetTTLStorage()
			err = ttlStorage.SetTTL(expiredKey, pastTime)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(activeKey, futureTime)
			Expect(err).NotTo(HaveOccurred())

			originalTmpDir := tmpDir
			testStorage.Close()

			newStorage, err := storage.NewLMDBStorage(originalTmpDir)
			Expect(err).NotTo(HaveOccurred())
			defer newStorage.Close()

			_, err = newStorage.Get(expiredKey)
			Expect(err).To(Equal(storage.ErrKeyNotFound))

			retrievedValue, err := newStorage.Get(activeKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))
		})

		It("should handle TTL restoration with mixed key states", func() {
			keys := [][]byte{
				[]byte("mixed_expired_1"),
				[]byte("mixed_active_1"),
				[]byte("mixed_persistent_1"),
				[]byte("mixed_expired_2"),
				[]byte("mixed_active_2"),
			}
			value := []byte("mixed_value")

			for _, key := range keys {
				err := testStorage.Set(key, value)
				Expect(err).NotTo(HaveOccurred())
			}

			pastTime := time.Now().Unix() - 100
			futureTime := time.Now().Unix() + 3600

			ttlStorage := testStorage.GetTTLStorage()
			err := ttlStorage.SetTTL(keys[0], pastTime)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[1], futureTime)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[3], pastTime)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(keys[4], futureTime)
			Expect(err).NotTo(HaveOccurred())

			originalTmpDir := tmpDir
			testStorage.Close()

			newStorage, err := storage.NewLMDBStorage(originalTmpDir)
			Expect(err).NotTo(HaveOccurred())
			defer newStorage.Close()

			_, err = newStorage.Get(keys[0])
			Expect(err).To(Equal(storage.ErrKeyNotFound))

			retrievedValue, err := newStorage.Get(keys[1])
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))

			retrievedValue, err = newStorage.Get(keys[2])
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))

			_, err = newStorage.Get(keys[3])
			Expect(err).To(Equal(storage.ErrKeyNotFound))

			retrievedValue, err = newStorage.Get(keys[4])
			Expect(err).NotTo(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))
		})
	})

	Describe("TTL integration with storage operations", func() {
		It("should maintain TTL consistency during concurrent operations", func() {
			baseKey := "concurrent_ttl_"
			value := []byte("concurrent_value")
			keyCount := 10

			for i := 0; i < keyCount; i++ {
				key := []byte(baseKey + string(rune('0'+i)))
				err := testStorage.Set(key, value)
				Expect(err).NotTo(HaveOccurred())

				result, err := ttlManager.SetExpire(key, int64(3600+i))
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(storage.ExpireSuccess))
			}

			for i := 0; i < keyCount; i++ {
				key := []byte(baseKey + string(rune('0'+i)))

				retrievedValue, err := testStorage.Get(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(retrievedValue).To(Equal(value))

				ttl, err := ttlManager.GetTTL(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(ttl).To(BeNumerically(">", int64(3500+i)))
			}

			deletedKeys := [][]byte{
				[]byte(baseKey + "0"),
				[]byte(baseKey + "2"),
				[]byte(baseKey + "4"),
			}

			deleted, err := testStorage.Del(deletedKeys...)
			Expect(err).NotTo(HaveOccurred())
			Expect(deleted).To(Equal(3))

			for _, key := range deletedKeys {
				_, err := testStorage.Get(key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))

				ttl, err := ttlManager.GetTTL(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(ttl).To(Equal(int64(storage.TTLNotFound)))
			}
		})

		It("should handle TTL operations on non-existent keys", func() {
			nonExistentKey := []byte("non_existent_ttl_key")

			result, err := ttlManager.SetExpire(nonExistentKey, 3600)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))

			ttl, err := ttlManager.GetTTL(nonExistentKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(int64(storage.TTLNotFound)))

			persistResult, err := ttlManager.Persist(nonExistentKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(persistResult).To(Equal(storage.PersistFailure))
		})
	})
})
