package storage

import (
	"os"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("TTL Persistence Property Tests", func() {
	var (
		tempDir string
		storage *LMDBStorage
	)

	BeforeEach(func() {
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tempDir, err = os.MkdirTemp("", "keyp-ttl-persistence-test")
		Expect(err).NotTo(HaveOccurred())

		storage, err = NewLMDBStorage(tempDir)
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if storage != nil {
			storage.Close()
		}
		os.RemoveAll(tempDir)
	})

	Describe("Property 7: TTL Setting and Querying", func() {
		It("should maintain TTL consistency during operations", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("TTL operations are consistent", prop.ForAll(
				func(keyData []byte, seconds int64) bool {
					if IsEmpty(keyData) || exceedsLimit(keyData) {
						return true
					}

					if seconds <= 0 || seconds > 300 {
						return true
					}

					key := make([]byte, len(keyData))
					copy(key, keyData)
					value := []byte("test-value")

					err := storage.Set(key, value)
					if HasError(err) {
						return false
					}

					ttlManager := storage.GetTTLManager()
					result, err := ttlManager.SetExpire(key, seconds)
					if HasError(err) || result != ExpireSuccess {
						return false
					}

					ttl, err := ttlManager.GetTTL(key)
					if HasError(err) || ttl <= 0 {
						return false
					}

					pttl, err := ttlManager.GetPTTL(key)
					if HasError(err) || pttl <= 0 {
						return false
					}

					return ttl > 0 && pttl > 0 && pttl >= ttl*1000

				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 300),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("Property 8: Expiration Behavior", func() {
		It("should handle expiration correctly", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 5

			properties := gopter.NewProperties(parameters)

			properties.Property("expired keys behave correctly", prop.ForAll(
				func(keyData []byte) bool {
					if IsEmpty(keyData) || exceedsLimit(keyData) {
						return true
					}

					key := make([]byte, len(keyData))
					copy(key, keyData)
					value := []byte("test-value")

					err := storage.Set(key, value)
					if HasError(err) {
						return false
					}

					ttlManager := storage.GetTTLManager()
					result, err := ttlManager.SetExpire(key, 1)
					if HasError(err) || result != ExpireSuccess {
						return false
					}

					time.Sleep(1100 * time.Millisecond)

					expired, err := ttlManager.IsExpired(key)
					if HasError(err) {
						return false
					}

					return expired

				},
				gen.SliceOfN(10, gen.UInt8()),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("Property 9: Persist Operations", func() {
		It("should handle persist operations correctly", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("persist operations work correctly", prop.ForAll(
				func(keyData []byte, seconds int64) bool {
					if IsEmpty(keyData) || exceedsLimit(keyData) {
						return true
					}

					if seconds <= 0 || seconds > 300 {
						return true
					}

					key := make([]byte, len(keyData))
					copy(key, keyData)
					value := []byte("test-value")

					err := storage.Set(key, value)
					if HasError(err) {
						return false
					}

					ttlManager := storage.GetTTLManager()
					result, err := ttlManager.SetExpire(key, seconds)
					if HasError(err) || result != ExpireSuccess {
						return false
					}

					persistResult, err := ttlManager.Persist(key)
					if HasError(err) || persistResult != PersistSuccess {
						return false
					}

					ttl, err := ttlManager.GetTTL(key)
					if HasError(err) {
						return false
					}

					return ttl == TTLPersistent

				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 300),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})
})
