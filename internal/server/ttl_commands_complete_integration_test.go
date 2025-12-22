package server_test

import (
	"context"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/server"
)

var _ = Describe("TTL Commands Complete Integration Tests", func() {
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
		tmpDir, err = os.MkdirTemp("", "ttl-complete-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6384", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6384",
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

	Describe("Complete TTL command suite via Redis client", func() {
		It("should handle complete TTL workflow with all commands", func() {
			key := "complete_workflow_key"
			value := "complete_workflow_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			ttl := client.TTL(ctx, key)
			Expect(ttl.Err()).NotTo(HaveOccurred())
			Expect(ttl.Val()).To(Equal(-1 * time.Nanosecond))

			pttl := client.PTTL(ctx, key)
			Expect(pttl.Err()).NotTo(HaveOccurred())
			Expect(pttl.Val()).To(Equal(-1 * time.Nanosecond))

			expireResult := client.Expire(ctx, key, 120*time.Second)
			Expect(expireResult.Err()).NotTo(HaveOccurred())
			Expect(expireResult.Val()).To(BeTrue())

			ttlAfterExpire := client.TTL(ctx, key)
			Expect(ttlAfterExpire.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterExpire.Val().Seconds()).To(BeNumerically(">", 115))
			Expect(ttlAfterExpire.Val().Seconds()).To(BeNumerically("<=", 120))

			pttlAfterExpire := client.PTTL(ctx, key)
			Expect(pttlAfterExpire.Err()).NotTo(HaveOccurred())
			Expect(pttlAfterExpire.Val().Milliseconds()).To(BeNumerically(">", 115000))
			Expect(pttlAfterExpire.Val().Milliseconds()).To(BeNumerically("<=", 120000))

			futureTimestamp := time.Now().Unix() + 180
			expireatResult := client.ExpireAt(ctx, key, time.Unix(futureTimestamp, 0))
			Expect(expireatResult.Err()).NotTo(HaveOccurred())
			Expect(expireatResult.Val()).To(BeTrue())

			ttlAfterExpireat := client.TTL(ctx, key)
			Expect(ttlAfterExpireat.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterExpireat.Val().Seconds()).To(BeNumerically(">", 175))
			Expect(ttlAfterExpireat.Val().Seconds()).To(BeNumerically("<=", 180))

			persistResult := client.Persist(ctx, key)
			Expect(persistResult.Err()).NotTo(HaveOccurred())
			Expect(persistResult.Val()).To(BeTrue())

			ttlAfterPersist := client.TTL(ctx, key)
			Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterPersist.Val()).To(Equal(-1 * time.Nanosecond))

			pttlAfterPersist := client.PTTL(ctx, key)
			Expect(pttlAfterPersist.Err()).NotTo(HaveOccurred())
			Expect(pttlAfterPersist.Val()).To(Equal(-1 * time.Nanosecond))

			retrievedValue := client.Get(ctx, key)
			Expect(retrievedValue.Err()).NotTo(HaveOccurred())
			Expect(retrievedValue.Val()).To(Equal(value))
		})

		It("should handle TTL commands with various data types", func() {
			testCases := []struct {
				key   string
				value string
				ttl   int64
			}{
				{"string_key", "simple_string", 60},
				{"json_key", `{"name":"test","value":123}`, 120},
				{"binary_key", string([]byte{0x00, 0x01, 0x02, 0xFF}), 180},
				{"unicode_key", "测试数据🚀", 240},
				{"long_key", string(make([]byte, 500)), 300},
			}

			for _, tc := range testCases {
				err := client.Set(ctx, tc.key, tc.value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				expireResult := client.Expire(ctx, tc.key, time.Duration(tc.ttl)*time.Second)
				Expect(expireResult.Err()).NotTo(HaveOccurred())
				Expect(expireResult.Val()).To(BeTrue())

				ttl := client.TTL(ctx, tc.key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(tc.ttl-5)))
				Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(tc.ttl)))

				retrievedValue := client.Get(ctx, tc.key)
				Expect(retrievedValue.Err()).NotTo(HaveOccurred())
				Expect(retrievedValue.Val()).To(Equal(tc.value))
			}
		})

		It("should handle TTL commands with edge cases", func() {
			key := "edge_case_key"
			value := "edge_case_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			zeroTTLResult := client.Expire(ctx, key, 0)
			Expect(zeroTTLResult.Err()).NotTo(HaveOccurred())
			Expect(zeroTTLResult.Val()).To(BeTrue())

			getResult := client.Get(ctx, key)
			Expect(getResult.Err()).To(HaveOccurred())
			Expect(getResult.Err().Error()).To(ContainSubstring("nil"))

			err = client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			negativeTTLResult := client.Expire(ctx, key, -100*time.Second)
			Expect(negativeTTLResult.Err()).NotTo(HaveOccurred())
			Expect(negativeTTLResult.Val()).To(BeFalse())

			largeTTLResult := client.Expire(ctx, key, 365*24*time.Hour)
			Expect(largeTTLResult.Err()).NotTo(HaveOccurred())
			Expect(largeTTLResult.Val()).To(BeTrue())

			ttl := client.TTL(ctx, key)
			Expect(ttl.Err()).NotTo(HaveOccurred())
			Expect(ttl.Val().Seconds()).To(BeNumerically(">", 365*24*3600-300))

			pastTimestamp := time.Now().Unix() - 100
			pastExpireatResult := client.ExpireAt(ctx, key, time.Unix(pastTimestamp, 0))
			Expect(pastExpireatResult.Err()).NotTo(HaveOccurred())
			Expect(pastExpireatResult.Val()).To(BeFalse())

			futureTimestamp := time.Now().Unix() + 86400
			futureExpireatResult := client.ExpireAt(ctx, key, time.Unix(futureTimestamp, 0))
			Expect(futureExpireatResult.Err()).NotTo(HaveOccurred())
			Expect(futureExpireatResult.Val()).To(BeTrue())
		})
	})

	Describe("Redis protocol compatibility for all TTL commands", func() {
		It("should match Redis response formats exactly", func() {
			key := "protocol_test_key"
			value := "protocol_test_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			expireResult := client.Do(ctx, "EXPIRE", key, "60")
			Expect(expireResult.Err()).NotTo(HaveOccurred())
			Expect(expireResult.Val()).To(Equal(int64(1)))

			ttlResult := client.Do(ctx, "TTL", key)
			Expect(ttlResult.Err()).NotTo(HaveOccurred())
			ttlValue, ok := ttlResult.Val().(int64)
			Expect(ok).To(BeTrue())
			Expect(ttlValue).To(BeNumerically(">", 55))
			Expect(ttlValue).To(BeNumerically("<=", 60))

			pttlResult := client.Do(ctx, "PTTL", key)
			Expect(pttlResult.Err()).NotTo(HaveOccurred())
			pttlValue, ok := pttlResult.Val().(int64)
			Expect(ok).To(BeTrue())
			Expect(pttlValue).To(BeNumerically(">", 55000))
			Expect(pttlValue).To(BeNumerically("<=", 60000))

			futureTimestamp := time.Now().Unix() + 120
			expireatResult := client.Do(ctx, "EXPIREAT", key, strconv.FormatInt(futureTimestamp, 10))
			Expect(expireatResult.Err()).NotTo(HaveOccurred())
			Expect(expireatResult.Val()).To(Equal(int64(1)))

			persistResult := client.Do(ctx, "PERSIST", key)
			Expect(persistResult.Err()).NotTo(HaveOccurred())
			Expect(persistResult.Val()).To(Equal(int64(1)))

			ttlAfterPersist := client.Do(ctx, "TTL", key)
			Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterPersist.Val()).To(Equal(int64(-1)))

			pttlAfterPersist := client.Do(ctx, "PTTL", key)
			Expect(pttlAfterPersist.Err()).NotTo(HaveOccurred())
			Expect(pttlAfterPersist.Val()).To(Equal(int64(-1)))
		})

		It("should handle non-existent keys with correct return codes", func() {
			nonExistentKey := "non_existent_protocol_key"

			expireResult := client.Do(ctx, "EXPIRE", nonExistentKey, "60")
			Expect(expireResult.Err()).NotTo(HaveOccurred())
			Expect(expireResult.Val()).To(Equal(int64(0)))

			expireatResult := client.Do(ctx, "EXPIREAT", nonExistentKey, strconv.FormatInt(time.Now().Unix()+60, 10))
			Expect(expireatResult.Err()).NotTo(HaveOccurred())
			Expect(expireatResult.Val()).To(Equal(int64(0)))

			ttlResult := client.Do(ctx, "TTL", nonExistentKey)
			Expect(ttlResult.Err()).NotTo(HaveOccurred())
			Expect(ttlResult.Val()).To(Equal(int64(-2)))

			pttlResult := client.Do(ctx, "PTTL", nonExistentKey)
			Expect(pttlResult.Err()).NotTo(HaveOccurred())
			Expect(pttlResult.Val()).To(Equal(int64(-2)))

			persistResult := client.Do(ctx, "PERSIST", nonExistentKey)
			Expect(persistResult.Err()).NotTo(HaveOccurred())
			Expect(persistResult.Val()).To(Equal(int64(0)))
		})
	})

	Describe("Error messages and response formats match Redis", func() {
		It("should return proper error messages for invalid arguments", func() {
			errorTestCases := []struct {
				command     []string
				expectedErr string
			}{
				{[]string{"EXPIRE"}, "wrong number of arguments"},
				{[]string{"EXPIRE", "key"}, "wrong number of arguments"},
				{[]string{"EXPIRE", "key", "60", "extra"}, "wrong number of arguments"},
				{[]string{"EXPIRE", "key", "invalid"}, "not an integer"},
				{[]string{"EXPIREAT"}, "wrong number of arguments"},
				{[]string{"EXPIREAT", "key"}, "wrong number of arguments"},
				{[]string{"EXPIREAT", "key", "1234567890", "extra"}, "wrong number of arguments"},
				{[]string{"EXPIREAT", "key", "invalid"}, "not an integer"},
				{[]string{"TTL"}, "wrong number of arguments"},
				{[]string{"TTL", "key", "extra"}, "wrong number of arguments"},
				{[]string{"PTTL"}, "wrong number of arguments"},
				{[]string{"PTTL", "key", "extra"}, "wrong number of arguments"},
				{[]string{"PERSIST"}, "wrong number of arguments"},
				{[]string{"PERSIST", "key", "extra"}, "wrong number of arguments"},
			}

			for _, tc := range errorTestCases {
				args := make([]interface{}, len(tc.command))
				for i, arg := range tc.command {
					args[i] = arg
				}
				result := client.Do(ctx, args...)
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring(tc.expectedErr))
			}
		})

		It("should handle command case insensitivity", func() {
			key := "case_test_key"
			value := "case_test_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			caseVariations := [][]string{
				{"expire", key, "60"},
				{"EXPIRE", key, "60"},
				{"Expire", key, "60"},
				{"eXpIrE", key, "60"},
				{"expireat", key, strconv.FormatInt(time.Now().Unix()+60, 10)},
				{"EXPIREAT", key, strconv.FormatInt(time.Now().Unix()+60, 10)},
				{"ttl", key},
				{"TTL", key},
				{"pttl", key},
				{"PTTL", key},
				{"persist", key},
				{"PERSIST", key},
			}

			for _, cmdArgs := range caseVariations {
				args := make([]interface{}, len(cmdArgs))
				for i, arg := range cmdArgs {
					args[i] = arg
				}
				result := client.Do(ctx, args...)
				Expect(result.Err()).NotTo(HaveOccurred())
			}
		})
	})

	Describe("TTL commands with various data types and edge cases", func() {
		It("should handle TTL operations with special characters in keys", func() {
			specialKeys := []string{
				"key:with:colons",
				"key-with-dashes",
				"key_with_underscores",
				"key.with.dots",
				"key with spaces",
				"key@with#special$chars%",
				"🔑unicode🗝️key",
			}

			for _, key := range specialKeys {
				value := "special_value_for_" + key

				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				expireResult := client.Expire(ctx, key, 60*time.Second)
				Expect(expireResult.Err()).NotTo(HaveOccurred())
				Expect(expireResult.Val()).To(BeTrue())

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val().Seconds()).To(BeNumerically(">", 55))

				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).NotTo(HaveOccurred())
				Expect(retrievedValue.Val()).To(Equal(value))

				persistResult := client.Persist(ctx, key)
				Expect(persistResult.Err()).NotTo(HaveOccurred())
				Expect(persistResult.Val()).To(BeTrue())
			}
		})

		It("should handle concurrent TTL operations on the same key", func() {
			key := "concurrent_ttl_key"
			value := "concurrent_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			numGoroutines := 5
			done := make(chan bool, numGoroutines)

			for i := 0; i < numGoroutines; i++ {
				go func(id int) {
					defer GinkgoRecover()

					testClient := redis.NewClient(&redis.Options{
						Addr: "localhost:6384",
					})
					defer testClient.Close()

					Eventually(func() error {
						return testClient.Ping(ctx).Err()
					}, "2s", "50ms").Should(Succeed())

					ttlValue := int64(60 + id*10)
					expireResult := testClient.Expire(ctx, key, time.Duration(ttlValue)*time.Second)
					Expect(expireResult.Err()).NotTo(HaveOccurred())
					Expect(expireResult.Val()).To(BeTrue())

					ttl := testClient.TTL(ctx, key)
					Expect(ttl.Err()).NotTo(HaveOccurred())
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", 0))

					pttl := testClient.PTTL(ctx, key)
					Expect(pttl.Err()).NotTo(HaveOccurred())
					Expect(pttl.Val().Milliseconds()).To(BeNumerically(">", 0))

					done <- true
				}(i)
			}

			for i := 0; i < numGoroutines; i++ {
				Eventually(done).Should(Receive())
			}

			finalTTL := client.TTL(ctx, key)
			Expect(finalTTL.Err()).NotTo(HaveOccurred())
			Expect(finalTTL.Val().Seconds()).To(BeNumerically(">", 0))

			retrievedValue := client.Get(ctx, key)
			Expect(retrievedValue.Err()).NotTo(HaveOccurred())
			Expect(retrievedValue.Val()).To(Equal(value))
		})
	})
})
