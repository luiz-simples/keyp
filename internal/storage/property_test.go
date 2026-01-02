package storage_test

import (
	"context"
	"fmt"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Storage Property-Based Tests", Label("property"), func() {
	var (
		client     *storage.Client
		testDir    string
		ctx        context.Context
		properties *gopter.Properties
	)

	BeforeEach(func() {
		testDir = createUniqueTestDir("property")
		var err error
		client, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(client).NotTo(BeNil())

		ctx = context.WithValue(context.Background(), domain.DB, uint8(0))

		parameters := gopter.DefaultTestParameters()
		parameters.MinSuccessfulTests = 100
		parameters.MaxSize = 50
		properties = gopter.NewProperties(parameters)
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("Set-Get Property", func() {
		Context("when setting and getting values", func() {
			It("should satisfy: Set(k,v) then Get(k) returns v", func() {
				properties.Property("set then get returns same value", prop.ForAll(
					func(key, value []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						err := client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						result, err := client.Get(ctx, key)
						if err != nil {
							return false
						}

						return string(result) == string(value)
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Set-Overwrite Property", func() {
		Context("when overwriting values", func() {
			It("should satisfy: Set(k,v1) then Set(k,v2) then Get(k) returns v2", func() {
				properties.Property("overwrite returns latest value", prop.ForAll(
					func(key, value1, value2 []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						err := client.Set(ctx, key, value1)
						if err != nil {
							return false
						}

						err = client.Set(ctx, key, value2)
						if err != nil {
							return false
						}

						result, err := client.Get(ctx, key)
						if err != nil {
							return false
						}

						return string(result) == string(value2)
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Set-Delete Property", func() {
		Context("when setting and deleting", func() {
			It("should satisfy: Set(k,v) then Del(k) then Get(k) returns ErrKeyNotFound", func() {
				properties.Property("delete removes key", prop.ForAll(
					func(key, value []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						err := client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						deleted, err := client.Del(ctx, key)
						if err != nil || deleted != 1 {
							return false
						}

						_, err = client.Get(ctx, key)
						return err == storage.ErrKeyNotFound
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Delete Non-Existent Property", func() {
		Context("when deleting non-existent keys", func() {
			It("should satisfy: Del(non-existent-key) returns 0 deleted", func() {
				properties.Property("delete non-existent returns zero", prop.ForAll(
					func(key []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						client.Del(ctx, key)

						deleted, err := client.Del(ctx, key)
						return err == nil && deleted == 0
					},
					gen.SliceOfN(10, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Multiple Delete Property", func() {
		Context("when deleting multiple keys", func() {
			It("should satisfy: Del(k1,k2,k3) returns count of existing keys", func() {
				properties.Property("multiple delete returns correct count", prop.ForAll(
					func(keys [][]byte) bool {
						if len(keys) == 0 {
							return true
						}

						uniqueKeys := make(map[string][]byte)
						for i, key := range keys {
							if len(key) == 0 {
								key = []byte(fmt.Sprintf("key-%d", i))
							}
							keyStr := string(key)
							uniqueKeys[keyStr] = key
						}

						keySlice := make([][]byte, 0, len(uniqueKeys))
						for _, key := range uniqueKeys {
							keySlice = append(keySlice, key)
						}

						if len(keySlice) == 0 {
							return true
						}

						expectedDeleted := len(keySlice) / 2
						for i := range expectedDeleted {
							err := client.Set(ctx, keySlice[i], []byte("value"))
							if err != nil {
								return false
							}
						}

						deleted, err := client.Del(ctx, keySlice...)
						return err == nil && int(deleted) == expectedDeleted
					},
					gen.SliceOfN(5, gen.SliceOfN(8, gen.UInt8())),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("TTL Property", func() {
		Context("when setting TTL", func() {
			It("should satisfy: Expire(k,ttl) then TTL(k) returns value > 0", func() {
				properties.Property("TTL returns positive value after expire", prop.ForAll(
					func(key, value []byte, ttlSeconds uint32) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}
						if ttlSeconds == 0 {
							ttlSeconds = 1
						}
						if ttlSeconds > 3600 {
							ttlSeconds = 3600
						}

						err := client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						client.Expire(ctx, key, ttlSeconds)
						retrievedTTL := client.TTL(ctx, key)

						return retrievedTTL > 0
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
					gen.UInt32Range(1, 3600),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Persist Property", func() {
		Context("when persisting keys", func() {
			It("should satisfy: Expire(k,ttl) then Persist(k) then TTL(k) returns 0", func() {
				properties.Property("persist removes TTL", prop.ForAll(
					func(key, value []byte, ttlSeconds uint32) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}
						if ttlSeconds == 0 {
							ttlSeconds = 1
						}
						if ttlSeconds > 3600 {
							ttlSeconds = 3600
						}

						err := client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						client.Expire(ctx, key, ttlSeconds)
						client.Persist(ctx, key)
						retrievedTTL := client.TTL(ctx, key)

						return retrievedTTL == 0
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
					gen.UInt32Range(1, 3600),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Database Isolation Property", func() {
		Context("when using different databases", func() {
			It("should satisfy: Set(db0,k,v1) and Set(db1,k,v2) maintain isolation", func() {
				properties.Property("database isolation maintained", prop.ForAll(
					func(key, value1, value2 []byte, db1, db2 uint8) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}
						if db1 == db2 {
							db2 = db1 + 1
						}
						if db1 > 99 {
							db1 = 99
						}
						if db2 > 99 {
							db2 = 99
						}

						ctx1 := context.WithValue(context.Background(), domain.DB, db1)
						ctx2 := context.WithValue(context.Background(), domain.DB, db2)

						err := client.Set(ctx1, key, value1)
						if err != nil {
							return false
						}

						err = client.Set(ctx2, key, value2)
						if err != nil {
							return false
						}

						result1, err := client.Get(ctx1, key)
						if err != nil {
							return false
						}

						result2, err := client.Get(ctx2, key)
						if err != nil {
							return false
						}

						return string(result1) == string(value1) && string(result2) == string(value2)
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
					gen.UInt8Range(0, 99),
					gen.UInt8Range(0, 99),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})

	Describe("Idempotency Properties", func() {
		Context("when performing idempotent operations", func() {
			It("should satisfy: Set(k,v) twice has same effect as Set(k,v) once", func() {
				properties.Property("set is idempotent", prop.ForAll(
					func(key, value []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						err := client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						result1, err := client.Get(ctx, key)
						if err != nil {
							return false
						}

						err = client.Set(ctx, key, value)
						if err != nil {
							return false
						}

						result2, err := client.Get(ctx, key)
						if err != nil {
							return false
						}

						return string(result1) == string(result2) && string(result1) == string(value)
					},
					gen.SliceOfN(10, gen.UInt8()),
					gen.SliceOfN(20, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})

			It("should satisfy: Del(non-existent-key) multiple times returns 0", func() {
				properties.Property("delete non-existent is idempotent", prop.ForAll(
					func(key []byte) bool {
						if len(key) == 0 {
							key = []byte("empty-key")
						}

						client.Del(ctx, key)

						deleted1, err := client.Del(ctx, key)
						if err != nil || deleted1 != 0 {
							return false
						}

						deleted2, err := client.Del(ctx, key)
						return err == nil && deleted2 == 0
					},
					gen.SliceOfN(10, gen.UInt8()),
				))

				Expect(properties.Run(gopter.ConsoleReporter(false))).To(BeTrue())
			})
		})
	})
})
