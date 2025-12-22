package storage_test

import (
	"os"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("LMDB Property Tests", func() {
	var (
		lmdbStorage *storage.LMDBStorage
		tmpDir      string
	)

	BeforeEach(func() {
		// Set test mode to disable logging during tests
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "lmdb-property-test-*")
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

	Describe("Set and Get Properties", func() {
		It("should satisfy set-get roundtrip property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("set-get roundtrip", prop.ForAll(
				func(key, value []byte) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					err := lmdbStorage.Set(key, value)
					if err != nil {
						return false
					}

					retrieved, err := lmdbStorage.Get(key)
					if err != nil {
						return false
					}

					return string(retrieved) == string(value)
				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.SliceOfN(100, gen.UInt8()),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})

		It("should satisfy key validation properties", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("empty keys should fail", prop.ForAll(
				func(value []byte) bool {
					err := lmdbStorage.Set([]byte{}, value)
					return err == storage.ErrEmptyKey
				},
				gen.SliceOfN(10, gen.UInt8()),
			))

			properties.Property("oversized keys should fail", prop.ForAll(
				func(value []byte) bool {
					largeKey := make([]byte, storage.MaxKeySize+1)
					err := lmdbStorage.Set(largeKey, value)
					return err == storage.ErrKeyTooLarge
				},
				gen.SliceOfN(10, gen.UInt8()),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("Delete Properties", func() {
		It("should satisfy delete count property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("delete returns correct count", prop.ForAll(
				func(keys [][]byte) bool {
					validKeys := make([][]byte, 0)
					expectedCount := 0

					for _, key := range keys {
						if len(key) == 0 || len(key) > storage.MaxKeySize {
							continue
						}

						validKeys = append(validKeys, key)
						err := lmdbStorage.Set(key, []byte("test"))
						if err != nil {
							return false
						}
						expectedCount++
					}

					if len(validKeys) == 0 {
						return true
					}

					deleted, err := lmdbStorage.Del(validKeys...)
					if err != nil {
						return false
					}

					return deleted == expectedCount
				},
				gen.SliceOfN(5, gen.SliceOfN(10, gen.UInt8())),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})

		It("should satisfy delete idempotency property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("delete is idempotent", prop.ForAll(
				func(key []byte) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					err := lmdbStorage.Set(key, []byte("test"))
					if err != nil {
						return false
					}

					firstDel, err := lmdbStorage.Del(key)
					if err != nil || firstDel != 1 {
						return false
					}

					secondDel, err := lmdbStorage.Del(key)
					if err != nil || secondDel != 0 {
						return false
					}

					return true
				},
				gen.SliceOfN(10, gen.UInt8()),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("Get Properties", func() {
		It("should satisfy non-existent key property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("get non-existent key returns error", prop.ForAll(
				func(key []byte) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					_, err := lmdbStorage.Get(key)
					return err == storage.ErrKeyNotFound
				},
				gen.SliceOfN(10, gen.UInt8()),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})
})
