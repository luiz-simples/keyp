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
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal(tc.success))

					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))

					ttl := client.TTL(ctx, tc.key)
					Expect(ttl.Err()).NotTo(HaveOccurred())
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(tc.ttl-5)))
					Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(tc.ttl)))
				}
			})

			It("should fail for non-existent keys", func() {

				result := client.Expire(ctx, "non:existent:key", 3600*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)
			})

			It("should handle negative TTL values", func() {

				key := "negative:ttl:key"
				value := "test:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, -100*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should handle large TTL values", func() {

				key := "large:ttl:key"
				value := "test:value"
				largeTTL := int64(365 * 24 * 3600) // 1 year

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(largeTTL)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(largeTTL-10)))
			})
		})

		Describe("EXPIREAT command via go-redis", func() {
			It("should set absolute expiration time with Unix timestamps", func() {

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
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					result := client.ExpireAt(ctx, tc.key, time.Unix(tc.timestamp, 0))
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(Equal(tc.success))

					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))

					ttl := client.TTL(ctx, tc.key)
					Expect(ttl.Err()).NotTo(HaveOccurred())
					expectedTTL := tc.timestamp - time.Now().Unix()
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(expectedTTL-5)))
					Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(expectedTTL+1)))
				}
			})

			It("should fail for non-existent keys", func() {

				futureTime := time.Now().Add(1 * time.Hour)
				result := client.ExpireAt(ctx, "non:existent:key", futureTime)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)
			})

			It("should fail for past timestamps", func() {

				key := "past:timestamp:key"
				value := "test:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				pastTime := time.Now().Add(-1 * time.Hour)
				result := client.ExpireAt(ctx, key, pastTime)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeFalse()) // Should return 0 (false)

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})
		})

		Describe("TTL/PTTL query commands via go-redis", func() {
			It("should return TTL in seconds for keys with expiration", func() {

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
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())

					ttlResult := client.TTL(ctx, tc.key)
					Expect(ttlResult.Err()).NotTo(HaveOccurred())

					actualTTL := int64(ttlResult.Val().Seconds())
					Expect(actualTTL).To(BeNumerically(">", tc.ttl-10))
					Expect(actualTTL).To(BeNumerically("<=", tc.ttl))

					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))
				}
			})

			It("should return PTTL in milliseconds with millisecond precision", func() {

				key := "pttl:precision:key"
				value := "precision:value"
				ttlSeconds := int64(30)

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				pttlResult := client.PTTL(ctx, key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())

				actualPTTL := pttlResult.Val().Milliseconds()
				expectedPTTLMin := (ttlSeconds - 5) * 1000
				expectedPTTLMax := ttlSeconds * 1000

				Expect(actualPTTL).To(BeNumerically(">", expectedPTTLMin))
				Expect(actualPTTL).To(BeNumerically("<=", expectedPTTLMax))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return -1 for persistent keys (no TTL)", func() {

				key := "persistent:ttl:key"
				value := "persistent:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				pttlResult := client.Do(ctx, "PTTL", key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())
				Expect(pttlResult.Val()).To(Equal(int64(-1)))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return -2 for non-existent keys", func() {

				nonExistentKey := "non:existent:ttl:key"

				ttlResult := client.Do(ctx, "TTL", nonExistentKey)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-2)))

				pttlResult := client.Do(ctx, "PTTL", nonExistentKey)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())
				Expect(pttlResult.Val()).To(Equal(int64(-2)))
			})

			It("should show TTL accuracy over time with multiple queries", func() {

				key := "time:accuracy:key"
				value := "time:value"
				initialTTL := int64(60) // 1 minute

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(initialTTL)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				firstTTL := client.TTL(ctx, key)
				Expect(firstTTL.Err()).NotTo(HaveOccurred())
				firstValue := int64(firstTTL.Val().Seconds())
				Expect(firstValue).To(BeNumerically(">", 0))
				Expect(firstValue).To(BeNumerically("<=", initialTTL))

				time.Sleep(100 * time.Millisecond)

				secondTTL := client.TTL(ctx, key)
				Expect(secondTTL.Err()).NotTo(HaveOccurred())
				secondValue := int64(secondTTL.Val().Seconds())
				Expect(secondValue).To(BeNumerically("<=", firstValue))

				firstPTTL := client.PTTL(ctx, key)
				Expect(firstPTTL.Err()).NotTo(HaveOccurred())
				firstPTTLValue := firstPTTL.Val().Milliseconds()
				Expect(firstPTTLValue).To(BeNumerically(">", 0))

				time.Sleep(50 * time.Millisecond)

				secondPTTL := client.PTTL(ctx, key)
				Expect(secondPTTL.Err()).NotTo(HaveOccurred())
				secondPTTLValue := secondPTTL.Val().Milliseconds()
				Expect(secondPTTLValue).To(BeNumerically("<=", firstPTTLValue))
			})

			It("should validate Redis protocol compatibility for TTL/PTTL commands", func() {

				key := "protocol:ttl:key"
				value := "protocol:value"
				ttl := int64(120)

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(ttl)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())

				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())

				ttlValue, ok := ttlResult.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(ttlValue).To(BeNumerically(">", 0))
				Expect(ttlValue).To(BeNumerically("<=", ttl))

				pttlResult := client.Do(ctx, "PTTL", key)
				Expect(pttlResult.Err()).NotTo(HaveOccurred())

				pttlValue, ok := pttlResult.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(pttlValue).To(BeNumerically(">", 0))
				Expect(pttlValue).To(BeNumerically(">=", ttlValue*1000-1000)) // Allow some tolerance
			})

			It("should return proper error messages for invalid TTL/PTTL arguments", func() {

				result := client.Do(ctx, "TTL")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "TTL", "key1", "key2")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

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
					err := client.Set(ctx, tc.key, tc.value, 0).Err()
					Expect(err).NotTo(HaveOccurred())

					result := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())

					ttlResult := client.Do(ctx, "TTL", tc.key)
					Expect(ttlResult.Err()).NotTo(HaveOccurred())
					ttlValue, ok := ttlResult.Val().(int64)
					Expect(ok).To(BeTrue())
					Expect(ttlValue).To(BeNumerically(">", 0))

					persistResult := client.Do(ctx, "PERSIST", tc.key)
					Expect(persistResult.Err()).NotTo(HaveOccurred())
					Expect(persistResult.Val()).To(Equal(int64(1))) // Should return 1 for success

					ttlAfterPersist := client.Do(ctx, "TTL", tc.key)
					Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
					Expect(ttlAfterPersist.Val()).To(Equal(int64(-1)))

					val, err := client.Get(ctx, tc.key).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(tc.value))
				}
			})

			It("should return 0 for persistent keys (already no TTL)", func() {

				key := "already:persistent:key"
				value := "persistent:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				persistResult := client.Do(ctx, "PERSIST", key)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(0))) // Should return 0 for no change

				ttlAfterPersist := client.Do(ctx, "TTL", key)
				Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
				Expect(ttlAfterPersist.Val()).To(Equal(int64(-1)))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return 0 for non-existent keys", func() {

				nonExistentKey := "non:existent:persist:key"

				persistResult := client.Do(ctx, "PERSIST", nonExistentKey)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(0))) // Should return 0 for failure

				_, err := client.Get(ctx, nonExistentKey).Result()
				Expect(err).To(Equal(redis.Nil))
			})

			It("should validate PERSIST idempotency via multiple calls", func() {

				key := "idempotent:persist:key"
				value := "idempotent:value"
				ttl := int64(120)

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, time.Duration(ttl)*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				persistResult1 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult1.Err()).NotTo(HaveOccurred())
				Expect(persistResult1.Val()).To(Equal(int64(1)))

				ttlResult := client.Do(ctx, "TTL", key)
				Expect(ttlResult.Err()).NotTo(HaveOccurred())
				Expect(ttlResult.Val()).To(Equal(int64(-1)))

				persistResult2 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult2.Err()).NotTo(HaveOccurred())
				Expect(persistResult2.Val()).To(Equal(int64(0)))

				persistResult3 := client.Do(ctx, "PERSIST", key)
				Expect(persistResult3.Err()).NotTo(HaveOccurred())
				Expect(persistResult3.Val()).To(Equal(int64(0)))

				ttlFinal := client.Do(ctx, "TTL", key)
				Expect(ttlFinal.Err()).NotTo(HaveOccurred())
				Expect(ttlFinal.Val()).To(Equal(int64(-1)))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should validate return codes and TTL removal behavior", func() {

				key := "return:codes:key"
				value := "return:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 60*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())

				ttlBefore := client.Do(ctx, "TTL", key)
				Expect(ttlBefore.Err()).NotTo(HaveOccurred())
				ttlValue, ok := ttlBefore.Val().(int64)
				Expect(ok).To(BeTrue())
				Expect(ttlValue).To(BeNumerically(">", 0))

				persistResult := client.Do(ctx, "PERSIST", key)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(Equal(int64(1)))

				ttlAfter := client.Do(ctx, "TTL", key)
				Expect(ttlAfter.Err()).NotTo(HaveOccurred())
				Expect(ttlAfter.Val()).To(Equal(int64(-1)))

				pttlAfter := client.Do(ctx, "PTTL", key)
				Expect(pttlAfter.Err()).NotTo(HaveOccurred())
				Expect(pttlAfter.Val()).To(Equal(int64(-1)))

				val, err := client.Get(ctx, key).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(value))
			})

			It("should return proper error messages for invalid PERSIST arguments", func() {

				result := client.Do(ctx, "PERSIST")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "PERSIST", "key1", "key2")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
			})

			It("should handle concurrent PERSIST operations", func() {

				numGoroutines := 3
				keysPerGoroutine := 2
				done := make(chan bool, numGoroutines)

				for g := 0; g < numGoroutines; g++ {
					go func(goroutineID int) {
						defer GinkgoRecover()

						testClient := redis.NewClient(&redis.Options{
							Addr: "localhost:6380",
						})
						defer testClient.Close()

						Eventually(func() error {
							return testClient.Ping(ctx).Err()
						}, "2s", "50ms").Should(Succeed())

						for k := 0; k < keysPerGoroutine; k++ {
							key := "concurrent:persist:" + string(rune(goroutineID)) + ":" + string(rune(k))
							value := "value:" + string(rune(goroutineID)) + ":" + string(rune(k))

							err := testClient.Set(ctx, key, value, 0).Err()
							Expect(err).NotTo(HaveOccurred())

							result := testClient.Expire(ctx, key, time.Duration(3600+k)*time.Second)
							Expect(result.Err()).NotTo(HaveOccurred())
							Expect(result.Val()).To(BeTrue())

							persistResult := testClient.Do(ctx, "PERSIST", key)
							Expect(persistResult.Err()).NotTo(HaveOccurred())
							Expect(persistResult.Val()).To(Equal(int64(1)))

							ttlResult := testClient.Do(ctx, "TTL", key)
							Expect(ttlResult.Err()).NotTo(HaveOccurred())
							Expect(ttlResult.Val()).To(Equal(int64(-1)))

							val, err := testClient.Get(ctx, key).Result()
							Expect(err).NotTo(HaveOccurred())
							Expect(val).To(Equal(value))
						}

						done <- true
					}(g)
				}

				for g := 0; g < numGoroutines; g++ {
					Eventually(done).Should(Receive())
				}
			})
		})

		Describe("Redis protocol compatibility", func() {
			It("should validate Redis protocol compatibility for TTL commands", func() {

				key := "protocol:test:key"
				value := "protocol:test:value"

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				expireResult := client.Do(ctx, "EXPIRE", key, "3600")
				Expect(expireResult.Err()).NotTo(HaveOccurred())
				Expect(expireResult.Val()).To(Equal(int64(1)))

				futureTimestamp := time.Now().Unix() + 7200
				expireatResult := client.Do(ctx, "EXPIREAT", key, futureTimestamp)
				Expect(expireatResult.Err()).NotTo(HaveOccurred())
				Expect(expireatResult.Val()).To(Equal(int64(1)))
			})

			It("should return proper error messages for invalid arguments", func() {

				result := client.Do(ctx, "EXPIRE", "key")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "EXPIRE", "key", "invalid")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("not an integer"))

				result = client.Do(ctx, "EXPIREAT", "key")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

				result = client.Do(ctx, "EXPIREAT", "key", "invalid")
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("not an integer"))
			})
		})

		Describe("Edge cases and error conditions", func() {
			It("should handle edge cases for EXPIRE/EXPIREAT commands", func() {

				key := "zero:ttl:key"
				value := "test:value"
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 0)
				Expect(result.Err()).NotTo(HaveOccurred())
				client.Get(ctx, key)
			})

			It("should handle concurrent EXPIRE/EXPIREAT operations", func() {

				numGoroutines := 5
				keysPerGoroutine := 3
				done := make(chan bool, numGoroutines)

				for g := 0; g < numGoroutines; g++ {
					go func(goroutineID int) {
						defer GinkgoRecover()

						testClient := redis.NewClient(&redis.Options{
							Addr: "localhost:6380",
						})
						defer testClient.Close()

						Eventually(func() error {
							return testClient.Ping(ctx).Err()
						}, "2s", "50ms").Should(Succeed())

						for k := 0; k < keysPerGoroutine; k++ {
							key := "concurrent:expire:" + string(rune(goroutineID)) + ":" + string(rune(k))
							value := "value:" + string(rune(goroutineID)) + ":" + string(rune(k))

							err := testClient.Set(ctx, key, value, 0).Err()
							Expect(err).NotTo(HaveOccurred())

							if k%2 == 0 {
								result := testClient.Expire(ctx, key, time.Duration(3600+k)*time.Second)
								Expect(result.Err()).NotTo(HaveOccurred())
								Expect(result.Val()).To(BeTrue())
							}
							if k%2 != 0 {
								futureTime := time.Now().Add(time.Duration(3600+k) * time.Second)
								result := testClient.ExpireAt(ctx, key, futureTime)
								Expect(result.Err()).NotTo(HaveOccurred())
								Expect(result.Val()).To(BeTrue())
							}

							val, err := testClient.Get(ctx, key).Result()
							Expect(err).NotTo(HaveOccurred())
							Expect(val).To(Equal(value))
						}

						done <- true
					}(g)
				}

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
