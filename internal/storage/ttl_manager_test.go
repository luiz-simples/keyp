package storage_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("TTL Manager Tests", func() {
	var (
		lmdbStorage *storage.LMDBStorage
		ttlManager  storage.TTLManager
		tmpDir      string
	)

	BeforeEach(func() {
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "ttl-manager-test-*")
		Expect(err).NotTo(HaveOccurred())

		lmdbStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		ttlManager = lmdbStorage.GetTTLManager()
	})

	AfterEach(func() {
		if lmdbStorage != nil {
			lmdbStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("SetExpire functionality", func() {
		It("should set TTL for existing keys", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))

			ttl, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically(">", 3500))
			Expect(ttl).To(BeNumerically("<=", 3600))
		})

		It("should fail for non-existent keys", func() {
			nonExistentKey := []byte("non:existent")

			result, err := ttlManager.SetExpire(nonExistentKey, 3600)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))
		})

		It("should fail for negative TTL", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.SetExpire(testKey, -100)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))
		})
	})

	Describe("SetExpireAt functionality", func() {
		It("should set absolute expiration time", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")
			futureTime := time.Now().Unix() + 7200

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.SetExpireAt(testKey, futureTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))

			ttl, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically(">", 7100))
			Expect(ttl).To(BeNumerically("<=", 7200))
		})

		It("should fail for past timestamps", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")
			pastTime := time.Now().Unix() - 100

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.SetExpireAt(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireFailure))
		})
	})

	Describe("GetTTL and GetPTTL functionality", func() {
		It("should return TTL in seconds and milliseconds", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			_, err = ttlManager.SetExpire(testKey, 1800)
			Expect(err).NotTo(HaveOccurred())

			ttlSeconds, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttlSeconds).To(BeNumerically(">", 1700))

			ttlMilliseconds, err := ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttlMilliseconds).To(BeNumerically(">", 1700000))
		})

		It("should return -1 for persistent keys", func() {
			testKey := []byte("persistent:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			ttl, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(int64(storage.TTLPersistent)))

			pttl, err := ttlManager.GetPTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl).To(Equal(int64(storage.TTLPersistent)))
		})

		It("should return -2 for non-existent keys", func() {
			nonExistentKey := []byte("non:existent")

			ttl, err := ttlManager.GetTTL(nonExistentKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(int64(storage.TTLNotFound)))

			pttl, err := ttlManager.GetPTTL(nonExistentKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl).To(Equal(int64(storage.TTLNotFound)))
		})
	})

	Describe("Persist functionality", func() {
		It("should remove TTL from keys", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			_, err = ttlManager.SetExpire(testKey, 3600)
			Expect(err).NotTo(HaveOccurred())

			ttl, err := ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(BeNumerically(">", 0))

			result, err := ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistSuccess))

			ttl, err = ttlManager.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(int64(storage.TTLPersistent)))
		})

		It("should fail for already persistent keys", func() {
			testKey := []byte("persistent:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			result, err := ttlManager.Persist(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(Equal(storage.PersistFailure))
		})
	})

	Describe("IsExpired functionality", func() {
		It("should detect expired keys", func() {
			testKey := []byte("test:key")
			testValue := []byte("test:value")
			pastTime := time.Now().Unix() - 100

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			ttlStorage := lmdbStorage.GetTTLStorage()
			err = ttlStorage.SetTTL(testKey, pastTime)
			Expect(err).NotTo(HaveOccurred())

			expired, err := ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeTrue())
		})

		It("should return false for persistent keys", func() {
			testKey := []byte("persistent:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(testKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			expired, err := ttlManager.IsExpired(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(expired).To(BeFalse())
		})
	})

	Describe("CleanupExpired functionality", func() {
		It("should remove expired keys and their TTL metadata", func() {
			expiredKey := []byte("expired:key")
			activeKey := []byte("active:key")
			testValue := []byte("test:value")

			err := lmdbStorage.Set(expiredKey, testValue)
			Expect(err).NotTo(HaveOccurred())
			err = lmdbStorage.Set(activeKey, testValue)
			Expect(err).NotTo(HaveOccurred())

			ttlStorage := lmdbStorage.GetTTLStorage()
			pastTime := time.Now().Unix() - 100
			futureTime := time.Now().Unix() + 3600

			err = ttlStorage.SetTTL(expiredKey, pastTime)
			Expect(err).NotTo(HaveOccurred())
			err = ttlStorage.SetTTL(activeKey, futureTime)
			Expect(err).NotTo(HaveOccurred())

			err = ttlManager.CleanupExpired()
			Expect(err).NotTo(HaveOccurred())

			_, err = lmdbStorage.Get(expiredKey)
			Expect(err).To(Equal(storage.ErrKeyNotFound))

			_, err = lmdbStorage.Get(activeKey)
			Expect(err).NotTo(HaveOccurred())

			_, err = ttlStorage.GetTTL(expiredKey)
			Expect(err).To(Equal(storage.ErrTTLNotFound))

			_, err = ttlStorage.GetTTL(activeKey)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
