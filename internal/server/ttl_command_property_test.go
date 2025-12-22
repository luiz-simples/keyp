package server_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/server"
)

var _ = Describe("TTL Command Property Tests", func() {
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
		tmpDir, err = os.MkdirTemp("", "ttl-property-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6383", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6383",
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

	Describe("Property 6: Command Validation Consistency", func() {
		It("should validate EXPIRE command arguments consistently", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.Rng.Seed(1234)
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("expire command validation", prop.ForAll(
				func(keyData []byte, seconds int64) bool {
					if len(keyData) == 0 || len(keyData) > 100 {
						return true
					}

					if seconds < -1000 || seconds > 86400 {
						return true
					}

					key := "prop_expire_" + string(keyData)
					value := "prop_value"

					err := client.Set(ctx, key, value, 0).Err()
					if err != nil {
						return false
					}

					result := client.Expire(ctx, key, time.Duration(seconds)*time.Second)
					if result.Err() != nil {
						return false
					}

					if seconds <= 0 {
						return !result.Val()
					}

					if !result.Val() {
						return false
					}

					ttl := client.TTL(ctx, key)
					if ttl.Err() != nil {
						return false
					}

					actualTTL := int64(ttl.Val().Seconds())
					return actualTTL > 0 && actualTTL <= seconds

				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(-100, 3600),
			))

			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})

		It("should validate EXPIREAT command arguments consistently", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.Rng.Seed(5678)
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("expireat command validation", prop.ForAll(
				func(keyData []byte, offsetSeconds int64) bool {
					if len(keyData) == 0 || len(keyData) > 100 {
						return true
					}

					if offsetSeconds < -3600 || offsetSeconds > 86400 {
						return true
					}

					key := "prop_expireat_" + string(keyData)
					value := "prop_value"
					timestamp := time.Now().Unix() + offsetSeconds

					err := client.Set(ctx, key, value, 0).Err()
					if err != nil {
						return false
					}

					result := client.ExpireAt(ctx, key, time.Unix(timestamp, 0))
					if result.Err() != nil {
						return false
					}

					if offsetSeconds <= 0 {
						return !result.Val()
					}

					if !result.Val() {
						return false
					}

					ttl := client.TTL(ctx, key)
					if ttl.Err() != nil {
						return false
					}

					actualTTL := int64(ttl.Val().Seconds())
					expectedTTL := timestamp - time.Now().Unix()
					return actualTTL > 0 && actualTTL <= expectedTTL+5

				},
				gen.SliceOfN(10, gen.UInt8()),
				gen.Int64Range(-100, 3600),
			))

			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})

		It("should validate TTL/PTTL command responses consistently", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.Rng.Seed(9012)
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("ttl query consistency", prop.ForAll(
				func(keyIndex int, seconds int64, hasTTL bool) bool {
					if seconds <= 0 || seconds > 3600 {
						return true
					}

					key := fmt.Sprintf("prop_ttl_key_%d", keyIndex)
					value := "prop_value"

					err := client.Set(ctx, key, value, 0).Err()
					if err != nil {
						return false
					}

					if hasTTL {
						result := client.Expire(ctx, key, time.Duration(seconds)*time.Second)
						if result.Err() != nil || !result.Val() {
							return false
						}

						ttlResult := client.Do(ctx, "TTL", key)
						if ttlResult.Err() != nil {
							return false
						}

						pttlResult := client.Do(ctx, "PTTL", key)
						if pttlResult.Err() != nil {
							return false
						}

						ttlValue, ok := ttlResult.Val().(int64)
						if !ok || ttlValue <= 0 {
							return false
						}

						pttlValue, ok := pttlResult.Val().(int64)
						if !ok || pttlValue <= 0 {
							return false
						}

						return pttlValue >= ttlValue*1000-1000 && pttlValue <= ttlValue*1000+1000
					}

					ttlResult := client.Do(ctx, "TTL", key)
					if ttlResult.Err() != nil {
						return false
					}

					pttlResult := client.Do(ctx, "PTTL", key)
					if pttlResult.Err() != nil {
						return false
					}

					ttlValue, ok := ttlResult.Val().(int64)
					if !ok {
						return false
					}

					pttlValue, ok := pttlResult.Val().(int64)
					if !ok {
						return false
					}

					return ttlValue == -1 && pttlValue == -1

				},
				gen.IntRange(1, 10000),
				gen.Int64Range(1, 300),
				gen.Bool(),
			))

			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})

		It("should validate PERSIST command behavior consistently", func() {
			parameters := gopter.DefaultTestParameters()
			parameters.Rng.Seed(3456)
			parameters.MinSuccessfulTests = 100

			properties := gopter.NewProperties(parameters)

			properties.Property("persist command consistency", prop.ForAll(
				func(keyIndex int, seconds int64, hasTTL bool) bool {
					if seconds <= 0 || seconds > 3600 {
						return true
					}

					key := fmt.Sprintf("prop_persist_key_%d", keyIndex)
					value := "prop_value"

					err := client.Set(ctx, key, value, 0).Err()
					if err != nil {
						return false
					}

					if hasTTL {
						result := client.Expire(ctx, key, time.Duration(seconds)*time.Second)
						if result.Err() != nil || !result.Val() {
							return false
						}
					}

					persistResult := client.Persist(ctx, key)
					if persistResult.Err() != nil {
						return false
					}

					expectedResult := hasTTL
					if persistResult.Val() != expectedResult {
						return false
					}

					ttlAfterPersist := client.Do(ctx, "TTL", key)
					if ttlAfterPersist.Err() != nil {
						return false
					}

					ttlValue, ok := ttlAfterPersist.Val().(int64)
					if !ok {
						return false
					}

					return ttlValue == -1

				},
				gen.IntRange(1, 10000),
				gen.Int64Range(1, 300),
				gen.Bool(),
			))

			Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
		})
	})

	Describe("Command Argument Validation", func() {
		It("should handle invalid argument counts consistently", func() {
			invalidCommands := [][]string{
				{"EXPIRE"},
				{"EXPIRE", "key"},
				{"EXPIRE", "key", "60", "extra"},
				{"EXPIREAT"},
				{"EXPIREAT", "key"},
				{"EXPIREAT", "key", "1234567890", "extra"},
				{"TTL"},
				{"TTL", "key", "extra"},
				{"PTTL"},
				{"PTTL", "key", "extra"},
				{"PERSIST"},
				{"PERSIST", "key", "extra"},
			}

			for _, cmdArgs := range invalidCommands {
				args := make([]interface{}, len(cmdArgs))
				for i, arg := range cmdArgs {
					args[i] = arg
				}
				result := client.Do(ctx, args...)
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
			}
		})

		It("should handle invalid numeric arguments consistently", func() {
			key := "numeric_test_key"
			value := "numeric_test_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			invalidNumericCommands := [][]string{
				{"EXPIRE", key, "not_a_number"},
				{"EXPIRE", key, "12.34"},
				{"EXPIRE", key, ""},
				{"EXPIREAT", key, "not_a_timestamp"},
				{"EXPIREAT", key, "12.34"},
				{"EXPIREAT", key, ""},
			}

			for _, cmdArgs := range invalidNumericCommands {
				args := make([]interface{}, len(cmdArgs))
				for i, arg := range cmdArgs {
					args[i] = arg
				}
				result := client.Do(ctx, args...)
				Expect(result.Err()).To(HaveOccurred())
				Expect(result.Err().Error()).To(ContainSubstring("not an integer"))
			}
		})

		It("should handle edge case numeric values consistently", func() {
			key := "edge_case_key"
			value := "edge_case_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			edgeCases := []struct {
				command   string
				value     string
				shouldErr bool
			}{
				{"EXPIRE", "0", false},
				{"EXPIRE", "-1", false},
				{"EXPIRE", strconv.FormatInt(int64(^uint(0)>>1), 10), false},
				{"EXPIREAT", "0", false},
				{"EXPIREAT", "-1", false},
				{"EXPIREAT", strconv.FormatInt(time.Now().Unix()+86400, 10), false},
			}

			for _, tc := range edgeCases {
				result := client.Do(ctx, tc.command, key, tc.value)
				if tc.shouldErr {
					Expect(result.Err()).To(HaveOccurred())
				} else {
					Expect(result.Err()).NotTo(HaveOccurred())
				}
			}
		})
	})
})
