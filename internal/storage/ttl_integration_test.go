package storage_test

import (
	"context"
	"os"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("TTL Integration Tests", func() {
	var (
		testStorage *storage.LMDBStorage
		ttlManager  storage.TTLManager
		tmpDir      string
		cancel      context.CancelFunc
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "ttl-integration-test-*")
		Expect(err).NotTo(HaveOccurred())

		testStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		ttlManager = testStorage.GetTTLManager()
		_, cancel = context.WithCancel(context.Background())
	})

	AfterEach(func() {
		cancel()
		if testStorage != nil {
			testStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("Property 4: Expiration Consistency", func() {
		It("should maintain expiration consistency across operations", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.Rng.Seed(1234)
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("expiration consistency", prop.ForAll(
				func(keyData []byte, seconds int64) bool {
					if storage.IsEmpty(keyData) || len(keyData) > 100 {
						return true
					}

					if seconds <= 0 || seconds > 3600 {
						return true
					}

					key := append([]byte("test_exp_"), keyData...)
					value := []byte("test_value")

					err := testStorage.Set(key, value)
					if storage.HasError(err) {
						return false
					}

					result, err := ttlManager.SetExpire(key, seconds)
					if storage.HasError(err) || result != storage.ExpireSuccess {
						return false
					}

					ttl, err := ttlManager.GetTTL(key)
					if storage.HasError(err) {
						return false
					}

					if ttl <= 0 || ttl > seconds {
						return false
					}

					expired, err := ttlManager.IsExpired(key)
					if storage.HasError(err) || expired {
						return false
					}

					retrievedValue, err := testStorage.Get(key)
					if storage.HasError(err) || string(retrievedValue) != string(value) {
						return false
					}

					testStorage.Del(key)
					return true
				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 300),
			))

			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("Automatic Expiration", func() {
		It("should automatically expire keys on access", func() {
			key := []byte("expire_test_key")
			value := []byte("expire_test_value")

			err := testStorage.Set(key, value)
			Expect(err).ToNot(HaveOccurred())

			result, err := ttlManager.SetExpire(key, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))

			retrievedValue, err := testStorage.Get(key)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))

			time.Sleep(2 * time.Second)

			retrievedValue, err = testStorage.Get(key)
			Expect(err).To(Equal(storage.ErrKeyNotFound))
			Expect(retrievedValue).To(BeNil())
		})

		It("should cleanup expired keys in background", func() {
			keys := [][]byte{
				[]byte("bg_cleanup_1"),
				[]byte("bg_cleanup_2"),
				[]byte("bg_cleanup_3"),
			}
			value := []byte("cleanup_value")

			for _, key := range keys {
				err := testStorage.Set(key, value)
				Expect(err).ToNot(HaveOccurred())

				result, err := ttlManager.SetExpire(key, 1)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(storage.ExpireSuccess))
			}

			time.Sleep(2 * time.Second)

			err := ttlManager.CleanupExpired()
			Expect(err).ToNot(HaveOccurred())

			for _, key := range keys {
				_, err := testStorage.Get(key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))
			}
		})

		It("should handle mixed expired and persistent keys", func() {
			expiredKey := []byte("will_expire")
			persistentKey := []byte("will_persist")
			value := []byte("mixed_value")

			err := testStorage.Set(expiredKey, value)
			Expect(err).ToNot(HaveOccurred())

			err = testStorage.Set(persistentKey, value)
			Expect(err).ToNot(HaveOccurred())

			result, err := ttlManager.SetExpire(expiredKey, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal(storage.ExpireSuccess))

			time.Sleep(2 * time.Second)

			_, err = testStorage.Get(expiredKey)
			Expect(err).To(Equal(storage.ErrKeyNotFound))

			retrievedValue, err := testStorage.Get(persistentKey)
			Expect(err).ToNot(HaveOccurred())
			Expect(retrievedValue).To(Equal(value))
		})
	})
})
