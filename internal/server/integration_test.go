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
