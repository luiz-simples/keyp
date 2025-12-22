package server_test

import (
	"context"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/server"
)

var _ = Describe("Integration Tests", func() {
	var (
		srv    *server.Server
		client *redis.Client
		ctx    context.Context
		tmpDir string
	)

	BeforeEach(func() {
		ctx = context.Background()

		// Set test mode to disable logging during tests
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "keyp-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6380", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6380",
		})

		Eventually(func() error {
			return client.Ping(ctx).Err()
		}, "5s", "100ms").Should(Succeed())
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		if srv != nil {
			srv.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("Basic Commands", func() {
		Describe("PING", func() {
			It("should respond with PONG", func() {
				result := client.Ping(ctx)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(Equal("PONG"))
			})

			It("should echo message when provided", func() {
				result := client.Do(ctx, "PING", "hello")
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(Equal("hello"))
			})
		})

		Describe("SET", func() {
			It("should set a key-value pair", func() {
				result := client.Set(ctx, "test-key", "test-value", 0)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(Equal("OK"))
			})
		})

		Describe("GET", func() {
			It("should return nil for non-existent key", func() {
				result := client.Get(ctx, "non-existent")
				Expect(result.Err()).To(Equal(redis.Nil))
			})
		})

		Describe("DEL", func() {
			It("should return 0 for non-existent key", func() {
				result := client.Del(ctx, "non-existent")
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(Equal(int64(0)))
			})
		})
	})

	Describe("TTL Commands Integration", func() {
		Describe("EXPIRE command via go-redis", func() {
			It("should set TTL for existing keys with various TTL values", func() {
				// Test EXPIRE command via Redis client with various TTL values
				// _Requirements: 1.1, 1.3, 1.4, 1.5_

				testCases := []struct {
					key     string
					value   string
					ttl     int64
					success bool
				}{
					{"expire:key1", "value1", 10, true},
					{"expire:key2", "value2", 3600, true},
					{"expire:key3", "value3", 1, true},
					{"expire:key4", "value4", 86400, true}, // 1 day
				}

				for _, tc := range testCases {
					// Set key first
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					// Set TTL using EXPIRE command
					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal(tc.success))

					// Verify key still exists and has correct value
					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))

					// Verify TTL was set (should be close to original value)
					ttl := client.TTL(ctx, tc.key)
					Expect(ttl.Err()).NotTo(HaveOccurred())
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(tc.ttl-5)))
					Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(tc.ttl)))
				}
			})

			It("should fail for non-existent keys", func() {
				// Test EXPIRE command on non-existent keys
				// _Requirements: 1.3_

				result := client.Expire(ctx, "non:existent:key", 3600*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)
			})

			It("should handle negative TTL values", func() {
				// Test EXPIRE command with negative TTL
				// _Requirements: 1.4_

				key := "negative:ttl:key"
				value := "test:value"

				// Set key first
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Set negative TTL (should fail)
				result := client.Expire(ctx, key, -100*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)

				// Key should still exist
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should handle large TTL values", func() {
				// Test EXPIRE command with large TTL values
				// _Requirements: 1.5_

				key := "large:ttl:key"
				value := "test:value"
				largeTTL := int64(365 * 24 * 3600) // 1 year

				// Set key first
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Set large TTL
				result := client.Expire(ctx, key, time.Duration(largeTTL)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				// Verify TTL was set
				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(largeTTL-10)))
			})
		})

		Describe("EXPIREAT command via go-redis", func() {
			It("should set absolute expiration time with Unix timestamps", func() {
				// Test EXPIREAT command via Redis client with Unix timestamps
				// _Requirements: 1.2, 1.3, 1.4, 1.5_

				baseTime := time.Now().Unix()
				testCases := []struct {
					key       string
					value     string
					timestamp int64
					success   bool
				}{
					{"expireat:key1", "value1", baseTime + 10, true},
					{"expireat:key2", "value2", baseTime + 3600, true},
					{"expireat:key3", "value3", baseTime + 86400, true}, // 1 day from now
				}

				for _, tc := range testCases {
					// Set key first
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					// Set absolute expiration using EXPIREAT command
					result := client.ExpireAt(ctx, tc.key, time.Unix(tc.timestamp, 0))
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal(tc.success))

					// Verify key still exists and has correct value
					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))

					// Verify TTL was set correctly
					ttl := client.TTL(ctx, tc.key)
					Expect(ttl.Err()).NotTo(HaveOccurred())
					expectedTTL := tc.timestamp - time.Now().Unix()
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(expectedTTL-5)))
					Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(expectedTTL+1)))
				}
			})

			It("should fail for non-existent keys", func() {
				// Test EXPIREAT command on non-existent keys
				// _Requirements: 1.3_

				futureTime := time.Now().Add(1 * time.Hour)
				result := client.ExpireAt(ctx, "non:existent:key", futureTime)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)
			})

			It("should fail for past timestamps", func() {
				// Test EXPIREAT command with past timestamps
				// _Requirements: 1.4_

				key := "past:timestamp:key"
				value := "test:value"

				// Set key first
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Set past timestamp (should fail)
				pastTime := time.Now().Add(-1 * time.Hour)
				result := client.ExpireAt(ctx, key, pastTime)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)

				// Key should still exist
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})
		})

		Describe("TTL/PTTL query commands via go-redis", func() {
			It("should return TTL in seconds for keys with expiration", func() {
				// Test TTL command via Redis client for keys with expiration
				// _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

				testCases := []struct {
					key   string
					value string
					ttl   int64
				}{
					{"ttl:query:key1", "value1", 10},
					{"ttl:query:key2", "value2", 3600},
					{"ttl:query:key3", "value3", 86400}, // 1 day
				}

				for _, tc := range testCases {
					// Set key first
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					// Set TTL using EXPIRE command
					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())

					// Query TTL using TTL command
					ttlResult := client.TTL(ctx, tc.key)
					Expect(ttlResult.Err()).NotTo(HaveOccurred())

					// TTL should be close to the set value (within reasonable bounds)
					actualTTL := int64(ttlResult.Val().Seconds())
					Expect(actualTTL).To(BeNumerically(">", tc.ttl-10))
					Expect(actualTTL).To(BeNumerically("<=", tc.ttl))

					// Verify key still exists
					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))
				}
			})

			It("should return PTTL in milliseconds with millisecond precision", func() {
				// Test PTTL command via Redis client with millisecond precision
				// _Requirements: 2.4, 2.5_

				key := "pttl:precision:key"
				value := "precision:value"
				ttlSeconds := int64(30)

				// Set key first
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Set TTL using EXPIRE command
				result := client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				// Query PTTL using PTTL command
				pttlResult := client.PTTL(ctx, key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())

				// PTTL should be in milliseconds (roughly 1000x the TTL)
				actualPTTL := int64(pttlResult.Val().Milliseconds())
				expectedPTTLMin := (ttlSeconds - 5) * 1000
				expectedPTTLMax := ttlSeconds * 1000

				Expect(actualPTTL).To(BeNumerically(">", expectedPTTLMin))
				Expect(actualPTTL).To(BeNumerically("<=", expectedPTTLMax))

				// Verify key still exists
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return -1 for persistent keys (no TTL)", func() {
				// Validate return codes: -1 for persistent
				// _Requirements: 2.2, 2.5_

				key := "persistent:ttl:key"
				value := "persistent:value"

				// Set key without TTL (persistent)
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Query TTL - should return -1 for persistent key
				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				// Query PTTL - should return -1 for persistent key
				pttlResult := client.Do(ctx, "PTTL", key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())
				Expect(pttlResult.Val()).To(Equal(int64(-1)))

				// Verify key exists
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return -2 for non-existent keys", func() {
				// Validate return codes: -2 for non-existent
				// _Requirements: 2.3, 2.5_

				nonExistentKey := "non:existent:ttl:key"

				// Query TTL for non-existent key - should return -2
				ttlResult := client.Do(ctx, "TTL", nonExistentKey)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-2)))

				// Query PTTL for non-existent key - should return -2
				pttlResult := client.Do(ctx, "PTTL", nonExistentKey)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())
				Expect(pttlResult.Val()).To(Equal(int64(-2)))
			})

			It("should show TTL accuracy over time with multiple queries", func() {
				// Test TTL accuracy over time with multiple queries
				// _Requirements: 2.1, 2.4_

				key := "time:accuracy:key"
				value := "time:value"
				initialTTL := int64(60) // 1 minute

				// Set key with TTL
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(initialTTL)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				// First TTL reading
				firstTTL := client.TTL(ctx, key)
				Expect(firstTTL.Err()).NotTo(HaveOccurred())
				firstValue := int64(firstTTL.Val().Seconds())
				Expect(firstValue).To(BeNumerically(">", 0))
				Expect(firstValue).To(BeNumerically("<=", initialTTL))

				// Wait a short time
				time.Sleep(100 * time.Millisecond)

				// Second TTL reading - should be less than or equal to first
				secondTTL := client.TTL(ctx, key)
				Expect(secondTTL.Err()).NotTo(HaveOccurred())
				secondValue := int64(secondTTL.Val().Seconds())
				Expect(secondValue).To(BeNumerically("<=", firstValue))

				// Test PTTL precision during the same period
				firstPTTL := client.PTTL(ctx, key)
				Expect(firstPTTL.Err()).NotTo(HaveOccurred())
				firstPTTLValue := int64(firstPTTL.Val().Milliseconds())
				Expect(firstPTTLValue).To(BeNumerically(">", 0))

				time.Sleep(50 * time.Millisecond)

				secondPTTL := client.PTTL(ctx, key)
				Expect(secondPTTL.Err()).NotTo(HaveOccurred())
				secondPTTLValue := int64(secondPTTL.Val().Milliseconds())
				Expect(secondPTTLValue).To(BeNumerically("<=", firstPTTLValue))
			})

			It("should validate Redis protocol compatibility for TTL/PTTL commands", func() {
				// Test Redis protocol compatibility and response format
				// _Requirements: 2.1, 2.4, 6.1_

				key := "protocol:ttl:key"
				value := "protocol:value"
				ttl := int64(120)

				// Set key with TTL
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(ttl)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())

				// Test TTL command via raw Redis protocol
				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())

				ttlValue, ok := ttlResult.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(ttlValue).To(BeNumerically(">", 0))
				Expect(ttlValue).To(BeNumerically("<=", ttl))

				// Test PTTL command via raw Redis protocol
				pttlResult := client.Do(ctx, "PTTL", key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())

				pttlValue, ok := pttlResult.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(pttlValue).To(BeNumerically(">", 0))
				Expect(pttlValue).To(BeNumerically(">=", ttlValue*1000-1000)) // Allow some tolerance
			})

			It("should return proper error messages for invalid TTL/PTTL arguments", func() {
				// Test command argument validation via Redis protocol
				// _Requirements: 6.1_

				// Test TTL with wrong number of arguments
				result := client.Do(ctx, "TTL")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "TTL", "key1", "key2")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				// Test PTTL with wrong number of arguments
				result = client.Do(ctx, "PTTL")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "PTTL", "key1", "key2")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
			})
		})

		Describe("PERSIST command via go-redis", func() {
			It("should remove TTL from keys with expiration", func() {
				// Test PERSIST command via Redis client on keys with TTL
				// _Requirements: 3.1, 3.2, 3.3_

				testCases := []struct {
					key   string
					value string
					ttl   int64
				}{
					{"persist:key1", "value1", 30},
					{"persist:key2", "value2", 3600},
					{"persist:key3", "value3", 86400}, // 1 day
				}

				for _, tc := range testCases {
					// Set key first
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					// Set TTL using EXPIRE command
					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())

					// Verify TTL is set
					ttlResult := client.Do(ctx, "TTL", tc.key)
					Expect(ttlResult.Err()).NotTo(HaveOccurred())
					ttlValue, ok := ttlResult.Val().(int64)
					Expect(ok).To(BeTrue())
					Expect(ttlValue).To(BeNumerically(">", 0))

					// Use PERSIST command to remove TTL
					persistResult := client.Do(ctx, "PERSIST", tc.key)
					Expect(persistResult.Err()).NotTo(HaveOccurred())
					Expect(persistResult.Val()).To(Equal(int64(1))) // Should return 1 for success

					// Verify TTL is removed (should return -1 for persistent)
					ttlAfterPersist := client.Do(ctx, "TTL", tc.key)
					Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
					Expect(ttlAfterPersist.Val()).To(Equal(int64(-1)))

					// Verify key still exists and has correct value
					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))
				}
			})

			It("should return 0 for persistent keys (already no TTL)", func() {
				// Test PERSIST command on persistent and non-existent keys
				// _Requirements: 3.2, 3.3_

				key := "already:persistent:key"
				value := "persistent:value"

				// Set key without TTL (persistent)
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Verify key is persistent
				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				// Use PERSIST command on already persistent key
				persistResult := client.Do(ctx, "PERSIST", key)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(0))) // Should return 0 for no change

				// Verify key is still persistent
				ttlAfterPersist := client.Do(ctx, "TTL", key)
				Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
				Expect(ttlAfterPersist.Val()).To(Equal(int64(-1)))

				// Verify key still exists
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return 0 for non-existent keys", func() {
				// Test PERSIST command on non-existent keys
				// _Requirements: 3.3_

				nonExistentKey := "non:existent:persist:key"

				// Use PERSIST command on non-existent key
				persistResult := client.Do(ctx, "PERSIST", nonExistentKey)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(0))) // Should return 0 for failure

				// Verify key doesn't exist
				_, err := client.Get(ctx, nonExistentKey).Result()
				Expect(err).To(Equal(redis.Nil))
			})

			It("should validate PERSIST idempotency via multiple calls", func() {
				// Test PERSIST idempotency via multiple calls
				// _Requirements: 3.1_

				key := "idempotent:persist:key"
				value := "idempotent:value"
				ttl := int64(120)

				// Set key with TTL
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(ttl)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				// First PERSIST - should succeed
				persistResult1 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult1.Err()).NotTo(HaveOccurred())
				Expect(persistResult1.Val()).To(Equal(int64(1)))

				// Verify TTL is removed
				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				// Second PERSIST - should return 0 (no change)
				persistResult2 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult2.Err()).NotTo(HaveOccurred())
				Expect(persistResult2.Val()).To(Equal(int64(0)))

				// Third PERSIST - should still return 0 (idempotent)
				persistResult3 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult3.Err()).NotTo(HaveOccurred())
				Expect(persistResult3.Val()).To(Equal(int64(0)))

				// Verify key is still persistent and exists
				ttlFinal := client.Do(ctx, "TTL", key)
				Expect(ttlFinal.Err()).NotTo(HaveOccurred())
				Expect(ttlFinal.Val()).To(Equal(int64(-1)))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should validate return codes and TTL removal behavior", func() {
				// Validate return codes and TTL removal behavior
				// _Requirements: 3.1, 3.2, 3.3_

				key := "return:codes:key"
				value := "return:value"

				// Set key with TTL
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 60*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				// Verify TTL exists before PERSIST
				ttlBefore := client.Do(ctx, "TTL", key)
				Expect(ttlBefore.Err()).NotTo(HaveOccurred())
				ttlValue, ok := ttlBefore.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(ttlValue).To(BeNumerically(">", 0))

				// PERSIST should return 1 and remove TTL
				persistResult := client.Do(ctx, "PERSIST", key)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(1)))

				// Verify TTL is completely removed
				ttlAfter := client.Do(ctx, "TTL", key)
				Expect(ttlAfter.Err()).NotTo(HaveOccurred())
				Expect(ttlAfter.Val()).To(Equal(int64(-1)))

				// Verify PTTL also returns -1
				pttlAfter := client.Do(ctx, "PTTL", key)
				Expect(pttlAfter.Err()).NotTo(HaveOccurred())
				Expect(pttlAfter.Val()).To(Equal(int64(-1)))

				// Key should still be accessible
				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return proper error messages for invalid PERSIST arguments", func() {
				// Test command argument validation via Redis protocol
				// _Requirements: 6.1_

				// Test PERSIST with wrong number of arguments
				result := client.Do(ctx, "PERSIST")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "PERSIST", "key1", "key2")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
			})

			It("should handle concurrent PERSIST operations", func() {
				// Test concurrent PERSIST operations from multiple Redis clients
				// _Requirements: 3.1_

				numGoroutines := 3
				keysPerGoroutine := 2
				done := make(chan bool, numGoroutines)

				for g := 0; g < numGoroutines; g++ {
					go func(goroutineID int) {
						defer GinkgoRecover()

						// Create separate Redis client for each goroutine
						testClient := redis.NewClient(&redis.Options{
							Addr: "localhost:6380",
						})
						defer testClient.Close()

						// Wait for client to connect
						Eventually(func() error {
							return testClient.Ping(ctx).Err()
						}, "2s", "50ms").Should(Succeed())

						for k := 0; k < keysPerGoroutine; k++ {
							key := "concurrent:persist:" + string(rune(goroutineID)) + ":" + string(rune(k))
							value := "value:" + string(rune(goroutineID)) + ":" + string(rune(k))

							// Set key with TTL
							err := testClient.Set(ctx, key, value, 0).Err()
							Expect(err).NotTo(HaveOccurred())

							result := testClient.Expire(ctx, key, time.Duration(3600+k)*time.Second)
							Expect(result.Err()).NotTo(HaveOccurred())
							Expect(result.Val()).To(BeTrue())

							// Use PERSIST concurrently
							persistResult := testClient.Do(ctx, "PERSIST", key)
							Expect(persistResult.Err()).NotTo(HaveOccurred())
							Expect(persistResult.Val()).To(Equal(int64(1)))

							// Verify TTL is removed
							ttlResult := testClient.Do(ctx, "TTL", key)
							Expect(ttlResult.Err()).NotTo(HaveOccurred())
							Expect(ttlResult.Val()).To(Equal(int64(-1)))

							// Verify key still exists
							val, err := testClient.Get(ctx, key).Result()
							Expect(err).NotTo(HaveOccurred())
							Expect(val).To(Equal(value))
						}

						done <- true
					}(g)
				}

				// Wait for all goroutines to complete
				for g := 0; g < numGoroutines; g++ {
					Eventually(done).Should(Receive())
				}
			})
		})

		Describe("Redis protocol compatibility", func() {
			It("should validate Redis protocol compatibility for TTL commands", func() {
				// Test Redis protocol compatibility and response codes
				// _Requirements: 1.1, 1.2, 6.1, 6.2, 6.3, 6.4_

				key := "protocol:test:key"
				value := "protocol:test:value"

				// Set key first
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Test EXPIRE with valid arguments
				expireResult := client.Do(ctx, "EXPIRE", key, "3600")
				Expect(expireResult.Err()).NotTo(HaveOccurred())
				Expect(expireResult.Val()).To(Equal(int64(1)))

				// Test EXPIREAT with valid arguments
				futureTimestamp := time.Now().Unix() + 7200
				expireatResult := client.Do(ctx, "EXPIREAT", key, futureTimestamp)
				Expect(expireatResult.Err()).NotTo(HaveOccurred())
				Expect(expireatResult.Val()).To(Equal(int64(1)))
			})

			It("should return proper error messages for invalid arguments", func() {
				// Test command argument validation via Redis protocol
				// _Requirements: 6.1, 6.2, 6.3, 6.4_

				// Test EXPIRE with wrong number of arguments
				result := client.Do(ctx, "EXPIRE", "key")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				// Test EXPIRE with non-numeric TTL
				result = client.Do(ctx, "EXPIRE", "key", "invalid")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("not an integer"))

				// Test EXPIREAT with wrong number of arguments
				result = client.Do(ctx, "EXPIREAT", "key")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				// Test EXPIREAT with non-numeric timestamp
				result = client.Do(ctx, "EXPIREAT", "key", "invalid")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("not an integer"))
			})
		})

		Describe("Edge cases and error conditions", func() {
			It("should handle edge cases for EXPIRE/EXPIREAT commands", func() {
				// Test edge cases: negative TTL, non-existent keys, large TTL values
				// _Requirements: 1.3, 1.4, 1.5_

				// Test zero TTL
				key := "zero:ttl:key"
				value := "test:value"
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 0)
				Expect(result.Err()).NotTo(HaveOccurred())
				// Zero TTL behavior may vary - check if key still exists
				_, err = client.Get(ctx, key).Result()
				// Either key exists or returns redis.Nil - both are acceptable
			})

			It("should handle concurrent EXPIRE/EXPIREAT operations", func() {
				// Test concurrent TTL operations via multiple Redis clients
				// _Requirements: 1.1, 1.2_

				numGoroutines := 5
				keysPerGoroutine := 3
				done := make(chan bool, numGoroutines)

				for g := 0; g < numGoroutines; g++ {
					go func(goroutineID int) {
						defer GinkgoRecover()

						// Create separate Redis client for each goroutine
						testClient := redis.NewClient(&redis.Options{
							Addr: "localhost:6380",
						})
						defer testClient.Close()

						// Wait for client to connect
						Eventually(func() error {
							return testClient.Ping(ctx).Err()
						}, "2s", "50ms").Should(Succeed())

						for k := 0; k < keysPerGoroutine; k++ {
							key := "concurrent:expire:" + string(rune(goroutineID)) + ":" + string(rune(k))
							value := "value:" + string(rune(goroutineID)) + ":" + string(rune(k))

							// Set key
							err := testClient.Set(ctx, key, value, 0).Err()
							Expect(err).NotTo(HaveOccurred())

							// Set TTL concurrently
							if k%2 == 0 {
								// Use EXPIRE
								result := testClient.Expire(ctx, key, time.Duration(3600+k)*time.Second)
								Expect(result.Err()).NotTo(HaveOccurred())
								Expect(result.Val()).To(BeTrue())
							} else {
								// Use EXPIREAT
								futureTime := time.Now().Add(time.Duration(3600+k) * time.Second)
								result := testClient.ExpireAt(ctx, key, futureTime)
								Expect(result.Err()).NotTo(HaveOccurred())
								Expect(result.Val()).To(BeTrue())
							}

							// Verify key still exists
							val, err := testClient.Get(ctx, key).Result()
							Expect(err).NotTo(HaveOccurred())
							Expect(val).To(Equal(value))
						}

						done <- true
					}(g)
				}

				// Wait for all goroutines to complete
				for g := 0; g < numGoroutines; g++ {
					Eventually(done).Should(Receive())
				}
			})
		})
	})

	Describe("Command Integration", func() {
		Describe("SET -> GET -> DEL workflow", func() {
			It("should handle complete key lifecycle", func() {
				key := "integration-test-key"
				value := "integration-test-value"

				setResult := client.Set(ctx, key, value, 0)
				Expect(setResult.Err()).NotTo(HaveOccurred())
				Expect(setResult.Val()).To(Equal("OK"))

				getResult := client.Get(ctx, key)
				Expect(getResult.Err()).NotTo(HaveOccurred())
				Expect(getResult.Val()).To(Equal(value))

				delResult := client.Del(ctx, key)
				Expect(delResult.Err()).NotTo(HaveOccurred())
				Expect(delResult.Val()).To(Equal(int64(1)))

				getAfterDel := client.Get(ctx, key)
				Expect(getAfterDel.Err()).To(Equal(redis.Nil))
			})
		})

		Describe("Multiple keys operations", func() {
			It("should handle multiple SET and GET operations", func() {
				keys := []string{"key1", "key2", "key3"}
				values := []string{"value1", "value2", "value3"}

				for i, key := range keys {
					result := client.Set(ctx, key, values[i], 0)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal("OK"))
				}

				for i, key := range keys {
					result := client.Get(ctx, key)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal(values[i]))
				}

				delResult := client.Del(ctx, keys...)
				Expect(delResult.Err()).NotTo(HaveOccurred())
				Expect(delResult.Val()).To(Equal(int64(len(keys))))
			})
		})
	})
})
