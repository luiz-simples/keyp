package storage_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Storage Unit Tests", Label("unit"), func() {
	var (
		client  *storage.Client
		testDir string
		ctx     context.Context
	)

	BeforeEach(func() {
		testDir = createUniqueTestDir("unit")
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

	Describe("Client Creation", func() {
		Context("when creating new client", func() {
			It("should create client successfully with valid directory", func() {
				tempDir := createUniqueTestDir("create")
				defer cleanupTestDir(tempDir)

				newClient, err := storage.NewClient(tempDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(newClient).NotTo(BeNil())
				newClient.Close()
			})

			It("should create directory if not exists", func() {
				nonExistentDir := createUniqueTestDir("nonexistent")
				defer cleanupTestDir(nonExistentDir)

				newClient, err := storage.NewClient(nonExistentDir)
				Expect(err).NotTo(HaveOccurred())
				Expect(newClient).NotTo(BeNil())
				newClient.Close()
			})
		})
	})

	Describe("Set Operation", func() {
		Context("when setting key-value pairs", func() {
			It("should set single key-value successfully", func() {
				key := []byte("test-key")
				value := []byte("test-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())
			})

			It("should set multiple different keys", func() {
				keys := [][]byte{
					[]byte("key1"),
					[]byte("key2"),
					[]byte("key3"),
				}
				values := [][]byte{
					[]byte("value1"),
					[]byte("value2"),
					[]byte("value3"),
				}

				for i := range keys {
					err := client.Set(ctx, keys[i], values[i])
					Expect(err).NotTo(HaveOccurred())
				}
			})

			It("should overwrite existing key", func() {
				key := []byte("overwrite-key")
				originalValue := []byte("original")
				newValue := []byte("updated")

				err := client.Set(ctx, key, originalValue)
				Expect(err).NotTo(HaveOccurred())

				err = client.Set(ctx, key, newValue)
				Expect(err).NotTo(HaveOccurred())

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(newValue))
			})

			It("should handle empty key", func() {
				key := []byte("")
				value := []byte("empty-key-value")

				err := client.Set(ctx, key, value)
				Expect(err).To(HaveOccurred())
			})

			It("should handle empty value", func() {
				key := []byte("empty-value-key")
				value := []byte("")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when context is cancelled", func() {
			It("should return context error", func() {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				key := []byte("cancelled-key")
				value := []byte("cancelled-value")

				err := client.Set(cancelCtx, key, value)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))
			})
		})
	})

	Describe("Get Operation", func() {
		Context("when getting existing keys", func() {
			It("should retrieve stored value", func() {
				key := []byte("get-key")
				expectedValue := []byte("get-value")

				err := client.Set(ctx, key, expectedValue)
				Expect(err).NotTo(HaveOccurred())

				result, err := client.Get(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result).To(Equal(expectedValue))
			})

			It("should retrieve multiple different values", func() {
				testData := map[string]string{
					"key1": "value1",
					"key2": "value2",
					"key3": "value3",
				}

				for k, v := range testData {
					err := client.Set(ctx, []byte(k), []byte(v))
					Expect(err).NotTo(HaveOccurred())
				}

				for k, expectedV := range testData {
					result, err := client.Get(ctx, []byte(k))
					Expect(err).NotTo(HaveOccurred())
					Expect(string(result)).To(Equal(expectedV))
				}
			})

			It("should handle empty key retrieval", func() {
				key := []byte("")

				result, err := client.Get(ctx, key)
				Expect(err).To(HaveOccurred())
				Expect(result).To(BeNil())
			})
		})

		Context("when getting non-existent keys", func() {
			It("should return ErrKeyNotFound", func() {
				key := []byte("non-existent-key")

				result, err := client.Get(ctx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(storage.ErrKeyNotFound))
				Expect(result).To(BeNil())
			})
		})

		Context("when context is cancelled", func() {
			It("should return context error", func() {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				key := []byte("cancelled-get-key")

				result, err := client.Get(cancelCtx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))
				Expect(result).To(BeNil())
			})
		})
	})

	Describe("Del Operation", func() {
		Context("when deleting existing keys", func() {
			It("should delete single key successfully", func() {
				key := []byte("delete-key")
				value := []byte("delete-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				deleted, err := client.Del(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(1)))

				_, err = client.Get(ctx, key)
				Expect(err).To(Equal(storage.ErrKeyNotFound))
			})

			It("should delete multiple keys successfully", func() {
				keys := [][]byte{
					[]byte("del-key1"),
					[]byte("del-key2"),
					[]byte("del-key3"),
				}

				for _, key := range keys {
					err := client.Set(ctx, key, []byte("value"))
					Expect(err).NotTo(HaveOccurred())
				}

				deleted, err := client.Del(ctx, keys...)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(3)))

				for _, key := range keys {
					_, err := client.Get(ctx, key)
					Expect(err).To(Equal(storage.ErrKeyNotFound))
				}
			})

			It("should handle mixed existing and non-existing keys", func() {
				existingKey := []byte("existing-key")
				nonExistingKey := []byte("non-existing-key")

				err := client.Set(ctx, existingKey, []byte("value"))
				Expect(err).NotTo(HaveOccurred())

				deleted, err := client.Del(ctx, existingKey, nonExistingKey)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(1)))
			})
		})

		Context("when deleting non-existent keys", func() {
			It("should return zero deleted count", func() {
				key := []byte("non-existent-delete-key")

				deleted, err := client.Del(ctx, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(0)))
			})
		})

		Context("when deleting with empty keys slice", func() {
			It("should return zero deleted count", func() {
				deleted, err := client.Del(ctx)
				Expect(err).NotTo(HaveOccurred())
				Expect(deleted).To(Equal(uint32(0)))
			})
		})

		Context("when context is cancelled", func() {
			It("should return context error", func() {
				cancelCtx, cancel := context.WithCancel(ctx)
				cancel()

				key := []byte("cancelled-delete-key")

				deleted, err := client.Del(cancelCtx, key)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(context.Canceled))
				Expect(deleted).To(Equal(uint32(0)))
			})
		})
	})

	Describe("TTL Operations", func() {
		Context("when setting and checking TTL", func() {
			It("should set TTL and retrieve correct value", func() {
				key := []byte("ttl-key")
				value := []byte("ttl-value")
				ttlSeconds := uint32(60)

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx, key, ttlSeconds)

				retrievedTTL := client.TTL(ctx, key)
				Expect(retrievedTTL).To(BeNumerically(">", 0))
				Expect(retrievedTTL).To(BeNumerically("<=", ttlSeconds+uint32(time.Now().Unix())))
			})

			It("should return zero TTL for non-existent key", func() {
				key := []byte("non-existent-ttl-key")

				ttl := client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))
			})

			It("should return zero TTL for key without TTL", func() {
				key := []byte("no-ttl-key")
				value := []byte("no-ttl-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				ttl := client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))
			})
		})

		Context("when persisting keys", func() {
			It("should remove TTL from key", func() {
				key := []byte("persist-key")
				value := []byte("persist-value")
				ttlSeconds := uint32(60)

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Expire(ctx, key, ttlSeconds)
				ttl := client.TTL(ctx, key)
				Expect(ttl).To(BeNumerically(">", 0))

				client.Persist(ctx, key)
				ttl = client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))
			})

			It("should handle persist on key without TTL", func() {
				key := []byte("no-ttl-persist-key")
				value := []byte("no-ttl-persist-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				client.Persist(ctx, key)
				ttl := client.TTL(ctx, key)
				Expect(ttl).To(Equal(uint32(0)))
			})
		})
	})

	Describe("Database Selection", func() {
		Context("when using different databases", func() {
			It("should isolate data between databases", func() {
				key := []byte("isolation-key")
				value1 := []byte("db0-value")
				value2 := []byte("db1-value")

				ctx0 := context.WithValue(context.Background(), domain.DB, uint8(0))
				ctx1 := context.WithValue(context.Background(), domain.DB, uint8(1))

				err := client.Set(ctx0, key, value1)
				Expect(err).NotTo(HaveOccurred())

				err = client.Set(ctx1, key, value2)
				Expect(err).NotTo(HaveOccurred())

				result0, err := client.Get(ctx0, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result0).To(Equal(value1))

				result1, err := client.Get(ctx1, key)
				Expect(err).NotTo(HaveOccurred())
				Expect(result1).To(Equal(value2))
			})
		})
	})

	Describe("Close Operation", func() {
		Context("when closing client", func() {
			It("should close without error", func() {
				tempDir := createUniqueTestDir("close")
				defer cleanupTestDir(tempDir)

				tempClient, err := storage.NewClient(tempDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(func() { tempClient.Close() }).NotTo(Panic())
			})
		})
	})
})
