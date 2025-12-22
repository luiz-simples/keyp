package storage_test

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("TTL Storage Property Tests", func() {
	var (
		lmdbStorage *storage.LMDBStorage
		ttlStorage  storage.TTLStorage
		tmpDir      string
	)

	BeforeEach(func() {
		var err error
		tmpDir, err = os.MkdirTemp("", "ttl-property-test-*")
		Expect(err).NotTo(HaveOccurred())

		lmdbStorage, err = storage.NewLMDBStorage(tmpDir)
		Expect(err).NotTo(HaveOccurred())

		ttlStorage = lmdbStorage.GetTTLStorage()
	})

	AfterEach(func() {
		if lmdbStorage != nil {
			lmdbStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("TTL Setting Consistency Properties", func() {
		It("should satisfy TTL setting consistency property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			// **Feature: ttl-commands, Property 1: TTL Setting Consistency**
			// **Validates: Requirements 1.1, 1.2**
			properties.Property("TTL setting consistency", prop.ForAll(
				func(key []byte, ttlSeconds int64) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					if ttlSeconds <= 0 {
						return true
					}

					testValue := []byte("test:value")

					// First set the key in storage (requirement for TTL operations)
					err := lmdbStorage.Set(key, testValue)
					if err != nil {
						return false
					}

					// Create TTL manager for testing
					ttlManager := storage.NewLMDBTTLManager(lmdbStorage)

					// Test EXPIRE command (SetExpire method)
					result, err := ttlManager.SetExpire(key, ttlSeconds)
					if err != nil {
						return false
					}

					if result != storage.ExpireSuccess {
						return false
					}

					// Verify TTL was set correctly within reasonable bounds
					actualTTL, err := ttlManager.GetTTL(key)
					if err != nil {
						return false
					}

					// TTL should be within reasonable bounds (allowing for execution time)
					// Should be positive and not exceed the original value
					if actualTTL <= 0 || actualTTL > ttlSeconds {
						return false
					}

					// TTL should be close to the set value (within 5 seconds tolerance for execution time)
					tolerance := int64(5)
					if ttlSeconds-actualTTL > tolerance {
						return false
					}

					// Test EXPIREAT command (SetExpireAt method) with future timestamp
					futureTimestamp := time.Now().Unix() + ttlSeconds
					result, err = ttlManager.SetExpireAt(key, futureTimestamp)
					if err != nil {
						return false
					}

					if result != storage.ExpireSuccess {
						return false
					}

					// Verify absolute expiration was set correctly
					actualTTL, err = ttlManager.GetTTL(key)
					if err != nil {
						return false
					}

					// TTL should be within reasonable bounds for absolute timestamp
					if actualTTL <= 0 || actualTTL > ttlSeconds {
						return false
					}

					// Clean up for next iteration
					lmdbStorage.Del(key)

					return true
				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 3600),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("TTL Persistence Round-trip Properties", func() {
		It("should satisfy TTL persistence round-trip property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			// **Feature: ttl-commands, Property 5: TTL Persistence Round-trip**
			// **Validates: Requirements 5.1, 5.2**
			properties.Property("TTL persistence round-trip", prop.ForAll(
				func(key []byte, ttlSeconds int64) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					if ttlSeconds <= 0 {
						return true
					}

					now := time.Now().Unix()
					expiresAt := now + ttlSeconds

					// Set TTL metadata
					err := ttlStorage.SetTTL(key, expiresAt)
					if err != nil {
						return false
					}

					// Verify TTL is set correctly before restart
					metadata, err := ttlStorage.GetTTL(key)
					if err != nil {
						return false
					}

					if metadata.ExpiresAt != expiresAt {
						return false
					}

					// Simulate system restart by closing and reopening storage
					originalTmpDir := tmpDir
					lmdbStorage.Close()

					// Reopen the same storage directory (simulating restart)
					newStorage, err := storage.NewLMDBStorage(originalTmpDir)
					if err != nil {
						return false
					}

					newTTLStorage := newStorage.GetTTLStorage()

					// Verify TTL metadata persisted across restart
					restoredMetadata, err := newTTLStorage.GetTTL(key)
					if err != nil {
						newStorage.Close()
						return false
					}

					// Verify all TTL information is preserved
					success := restoredMetadata.ExpiresAt == expiresAt &&
						string(restoredMetadata.Key) == string(key) &&
						restoredMetadata.CreatedAt > 0

					newStorage.Close()

					// Restore original storage for cleanup
					lmdbStorage, err = storage.NewLMDBStorage(originalTmpDir)
					if err != nil {
						return false
					}
					ttlStorage = lmdbStorage.GetTTLStorage()

					return success
				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 3600),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})

		It("should satisfy TTL removal property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("TTL removal consistency", prop.ForAll(
				func(key []byte, ttlSeconds int64) bool {
					if len(key) == 0 || len(key) > storage.MaxKeySize {
						return true
					}

					if ttlSeconds <= 0 {
						return true
					}

					now := time.Now().Unix()
					expiresAt := now + ttlSeconds

					err := ttlStorage.SetTTL(key, expiresAt)
					if err != nil {
						return false
					}

					err = ttlStorage.RemoveTTL(key)
					if err != nil {
						return false
					}

					_, err = ttlStorage.GetTTL(key)
					return err == storage.ErrTTLNotFound
				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(1, 3600),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})

		It("should satisfy expired keys query property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("expired keys query accuracy", prop.ForAll(
				func(keyCount int, randomSeed int64) bool {
					if keyCount <= 0 || keyCount > 8 {
						return true
					}

					// Use current time + random seed to ensure unique timestamps
					baseTime := time.Now().Unix() + randomSeed%1000000

					expiredCount := 0

					for i := 0; i < keyCount; i++ {
						// Use unique key with timestamp and random seed to avoid conflicts
						key := []byte(fmt.Sprintf("prop:key:%d:%d:%d", baseTime, randomSeed, i))

						if len(key) == 0 || len(key) > storage.MaxKeySize {
							continue
						}

						var expiresAt int64
						if i%2 == 0 {
							expiresAt = baseTime - 100
							expiredCount++
						} else {
							expiresAt = baseTime + 100
						}

						err := ttlStorage.SetTTL(key, expiresAt)
						if err != nil {
							return false
						}
					}

					expiredKeys, err := ttlStorage.GetExpiredKeys(baseTime)
					if err != nil {
						return false
					}

					// Count only keys from this test run
					actualExpiredCount := 0
					keyPrefix := fmt.Sprintf("prop:key:%d:%d:", baseTime, randomSeed)
					for _, expiredKey := range expiredKeys {
						if strings.HasPrefix(string(expiredKey), keyPrefix) {
							actualExpiredCount++
						}
					}

					return actualExpiredCount == expiredCount
				},
				gen.IntRange(1, 8),
				gen.Int64Range(1, 1000000),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})

		It("should satisfy batch removal property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("batch TTL removal consistency", prop.ForAll(
				func(keys [][]byte, ttlSeconds int64) bool {
					if ttlSeconds <= 0 {
						return true
					}

					validKeys := make([][]byte, 0)
					now := time.Now().Unix()
					expiresAt := now + ttlSeconds

					for _, key := range keys {
						if len(key) == 0 || len(key) > storage.MaxKeySize {
							continue
						}

						validKeys = append(validKeys, key)
						err := ttlStorage.SetTTL(key, expiresAt)
						if err != nil {
							return false
						}
					}

					if len(validKeys) == 0 {
						return true
					}

					err := ttlStorage.RemoveTTLBatch(validKeys)
					if err != nil {
						return false
					}

					for _, key := range validKeys {
						_, err := ttlStorage.GetTTL(key)
						if err != storage.ErrTTLNotFound {
							return false
						}
					}

					return true
				},
				gen.SliceOfN(5, gen.SliceOfN(8, gen.UInt8())),
				gen.Int64Range(1, 3600),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})

	Describe("TTL Validation Properties", func() {
		It("should satisfy TTL validation property", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.MinSuccessfulTests = 50

			properties := gopter.NewProperties(parameters)

			properties.Property("empty keys should fail TTL operations", prop.ForAll(
				func(ttlSeconds int64) bool {
					if ttlSeconds <= 0 {
						return true
					}

					now := time.Now().Unix()
					expiresAt := now + ttlSeconds

					err := ttlStorage.SetTTL([]byte{}, expiresAt)
					if err != storage.ErrEmptyKey {
						return false
					}

					_, err = ttlStorage.GetTTL([]byte{})
					if err != storage.ErrEmptyKey {
						return false
					}

					err = ttlStorage.RemoveTTL([]byte{})
					return err == storage.ErrEmptyKey
				},
				gen.Int64Range(1, 3600),
			))

			properties.Property("oversized keys should fail TTL operations", prop.ForAll(
				func(ttlSeconds int64) bool {
					if ttlSeconds <= 0 {
						return true
					}

					largeKey := make([]byte, storage.MaxKeySize+1)
					now := time.Now().Unix()
					expiresAt := now + ttlSeconds

					err := ttlStorage.SetTTL(largeKey, expiresAt)
					if err != storage.ErrKeyTooLarge {
						return false
					}

					_, err = ttlStorage.GetTTL(largeKey)
					if err != storage.ErrKeyTooLarge {
						return false
					}

					err = ttlStorage.RemoveTTL(largeKey)
					return err == storage.ErrKeyTooLarge
				},
				gen.Int64Range(1, 3600),
			))

			result := properties.Run(gopter.ConsoleReporter(false))
			Expect(result).To(BeTrue())
		})
	})
})
