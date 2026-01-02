package storage_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Storage Integration Tests", Label("integration"), func() {
	var (
		client  *storage.Client
		testDir string
		ctx     context.Context
	)

	BeforeEach(func() {
		testDir = createUniqueTestDir("integration")
		var err error
		client, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(client).NotTo(BeNil())

		ctx = context.WithValue(context.Background(), domain.DB, uint8(0))
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("Concurrent Operations", func() {
		Context("when performing concurrent sets", func() {
			It("should handle multiple goroutines setting different keys", func() {
				const numGoroutines = 10
				const keysPerGoroutine = 100

				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				for i := range numGoroutines {
					go func(goroutineID int) {
						defer wg.Done()
						defer GinkgoRecover()

						for j := range keysPerGoroutine {
							key := []byte(fmt.Sprintf("concurrent-key-%d-%d", goroutineID, j))
							value := []byte(fmt.Sprintf("concurrent-value-%d-%d", goroutineID, j))

							err := client.Set(ctx, key, value)
							Expect(err).NotTo(HaveOccurred())
						}
					}(i)
				}

				wg.Wait()

				for i := range numGoroutines {
					for j := range keysPerGoroutine {
						key := []byte(fmt.Sprintf("concurrent-key-%d-%d", i, j))
						expectedValue := []byte(fmt.Sprintf("concurrent-value-%d-%d", i, j))

						result, err := client.Get(ctx, key)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(Equal(expectedValue))
					}
				}
			})

			It("should handle concurrent operations on same key", func() {
				const numGoroutines = 50
				key := []byte("shared-key")

				var wg sync.WaitGroup
				wg.Add(numGoroutines)

				for i := range numGoroutines {
					go func(goroutineID int) {
						defer wg.Done()
						defer GinkgoRecover()

						value := []byte(fmt.Sprintf("value-%d", goroutineID))
						err := client.Set(ctx, key, value)
						Expect(err).NotTo(HaveOccurred())
					}(i)
				}

				wg.Wait()

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).NotTo(BeEmpty())
			})
		})

		Context("when performing mixed concurrent operations", func() {
			It("should handle concurrent set, get, and delete operations", func() {
				const numOperations = 100
				baseKey := "mixed-op-key"

				var wg sync.WaitGroup
				wg.Add(numOperations * 3)

				for i := range numOperations {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						key := []byte(fmt.Sprintf("%s-set-%d", baseKey, id))
						value := []byte(fmt.Sprintf("value-%d", id))
						err := client.Set(ctx, key, value)
						Expect(err).NotTo(HaveOccurred())
					}(i)
				}

				for i := range numOperations {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						key := []byte(fmt.Sprintf("%s-get-%d", baseKey, id))
						_, _ = client.Get(ctx, key)
					}(i)
				}

				for i := range numOperations {
					go func(id int) {
						defer wg.Done()
						defer GinkgoRecover()

						key := []byte(fmt.Sprintf("%s-del-%d", baseKey, id))
						_, _ = client.Del(ctx, key)
					}(i)
				}

				wg.Wait()
			})
		})
	})

	Describe("TTL Integration", func() {
		Context("when keys expire", func() {
			It("should set and retrieve TTL values", func() {
				key := []byte("ttl-test-key")
				value := []byte("ttl-test-value")
				ttlSeconds := uint32(3600)

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx, key, ttlSeconds)

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))

				ttl := client.TTL(ctx, key)
				Expect(ttl).To(BeNumerically(">", 0))
			})

			It("should handle TTL persistence", func() {
				key := []byte("persist-test-key")
				value := []byte("persist-test-value")
				ttlSeconds := uint32(3600)

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx, key, ttlSeconds)
				ttl := client.TTL(ctx, key)
				Expect(ttl).To(BeNumerically(">", 0))

				client.Persist(ctx, key)
				ttl = client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))
			})
		})

		Context("when persisting keys with TTL", func() {
			It("should remove TTL without deleting key", func() {
				key := []byte("persist-key")
				value := []byte("persist-value")
				ttlSeconds := uint32(3600)

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx, key, ttlSeconds)
				client.Persist(ctx, key)

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(value))

				ttl := client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))
			})
		})
	})

	Describe("Multi-Database Operations", func() {
		Context("when using multiple databases", func() {
			It("should maintain isolation between databases", func() {
				key := []byte("multi-db-key")
				value0 := []byte("db0-value")
				value1 := []byte("db1-value")
				value2 := []byte("db2-value")

				ctx0 := context.WithValue(context.Background(), domain.DB, uint8(0))
				ctx1 := context.WithValue(context.Background(), domain.DB, uint8(1))
				ctx2 := context.WithValue(context.Background(), domain.DB, uint8(2))

				err := client.Set(ctx0, key, value0)
				Expect(err).NotTo(HaveOccurred())

				err = client.Set(ctx1, key, value1)
				Expect(err).NotTo(HaveOccurred())

				err = client.Set(ctx2, key, value2)
				Expect(err).NotTo(HaveOccurred())

				result0, err := client.Get(ctx0, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result0).To(Equal(value0))

				result1, err := client.Get(ctx1, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1).To(Equal(value1))

				result2, err := client.Get(ctx2, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result2).To(Equal(value2))

				deleted, err := client.Del(ctx0, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(1)))

				_, err = client.Get(ctx0, key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))

				_, err = client.Get(ctx1, key)
				Expect(err).NotTo(HaveOccurred())

				_, err = client.Get(ctx2, key)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should handle TTL isolation between databases", func() {
				key := []byte("ttl-isolation-key")
				value := []byte("ttl-isolation-value")

				ctx0 := context.WithValue(context.Background(), domain.DB, uint8(0))
				ctx1 := context.WithValue(context.Background(), domain.DB, uint8(1))

				err := client.Set(ctx0, key, value)
				Expect(err).NotTo(HaveOccurred())

				err = client.Set(ctx1, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx0, key, 1)

				ttl0 := client.TTL(ctx0, key)
				Expect(ttl0).To(BeNumerically(">", 0))

				ttl1 := client.TTL(ctx1, key)
				Expect(ttl1).To(Equal(uint32(0)))

				time.Sleep(2 * time.Second)

				_, err = client.Get(ctx0, key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))

				result1, err := client.Get(ctx1, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1).To(Equal(value))
			})
		})
	})

	Describe("Large Data Operations", func() {
		Context("when handling large values", func() {
			It("should store and retrieve large values", func() {
				key := []byte("large-value-key")
				largeValue := make([]byte, 1024*1024)
				for i := range largeValue {
					largeValue[i] = byte(i % 256)
				}

				err := client.Set(ctx, key, largeValue)
				Expect(err).NotTo(HaveOccurred())

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(largeValue))
			})

			It("should handle many small keys", func() {
				const numKeys = 10000
				keyPrefix := "small-key-"
				value := []byte("small-value")

				for i := range numKeys {
					key := []byte(fmt.Sprintf("%s%d", keyPrefix, i))
					err := client.Set(ctx, key, value)
					Expect(err).NotTo(HaveOccurred())
				}

				for i := range numKeys {
					key := []byte(fmt.Sprintf("%s%d", keyPrefix, i))
					result, err := client.Get(ctx, key)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(value))
				}

				keys := make([][]byte, numKeys)
				for i := range numKeys {
					keys[i] = []byte(fmt.Sprintf("%s%d", keyPrefix, i))
				}

				deleted, err := client.Del(ctx, keys...)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(numKeys)))
			})
		})
	})

	Describe("Error Handling", func() {
		Context("when operations fail", func() {
			It("should handle context timeout", func() {
				key := []byte("timeout-key")
				value := []byte("timeout-value")

				timeoutCtx, cancel := context.WithTimeout(ctx, 1*time.Nanosecond)
				defer cancel()

				time.Sleep(1 * time.Millisecond)

				err := client.Set(timeoutCtx, key, value)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.DeadlineExceeded))
			})

			It("should handle context cancellation during operations", func() {
				key := []byte("cancel-key")
				value := []byte("cancel-value")

				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				err := client.Set(cancelCtx, key, value)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))

				_, err = client.Get(cancelCtx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))

				_, err = client.Del(cancelCtx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))
			})
		})
	})
})
