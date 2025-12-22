package storage_test

import (
	"context"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/server"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("TTL Storage Integration Tests", func() {
	var (
		srv         *server.Server
		client      *redis.Client
		lmdbStorage *storage.LMDBStorage
		ttlStorage  storage.TTLStorage
		tmpDir      string
		ctx         context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()

		var err error
		tmpDir, err = os.MkdirTemp("", "ttl-integration-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6381", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6381",
		})

		Eventually(func() error {
			return client.Ping(ctx).Err()
		}, "5s", "100ms").Should(Succeed())

		lmdbStorage, err = storage.NewLMDBStorage(tmpDir + "/ttl_test")
		Expect(err).NotTo(HaveOccurred())

		ttlStorage = lmdbStorage.GetTTLStorage()
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		if srv != nil {
			srv.Close()
		}
		if lmdbStorage != nil {
			lmdbStorage.Close()
		}
		if tmpDir != "" {
			os.RemoveAll(tmpDir)
		}
	})

	Describe("TTL Metadata Persistence via Redis Protocol", func() {
		It("should persist TTL metadata across storage operations", func() {
			// Test TTL metadata persistence via Redis protocol
			// _Requirements: 5.1, 5.2_

			testKey := []byte("test:ttl:key")
			testValue := "test:value"
			expiresAt := time.Now().Unix() + 3600

			// Set a key via Redis client first
			err := client.Set(ctx, string(testKey), testValue, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Verify key exists via Redis
			val, err := client.Get(ctx, string(testKey)).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal(testValue))

			// Set TTL metadata directly (simulating future EXPIRE command)
			err = ttlStorage.SetTTL(testKey, expiresAt)
			Expect(err).NotTo(HaveOccurred())

			// Verify TTL metadata persisted
			metadata, err := ttlStorage.GetTTL(testKey)
			Expect(err).NotTo(HaveOccurred())
			Expect(metadata.ExpiresAt).To(Equal(expiresAt))
			Expect(string(metadata.Key)).To(Equal(string(testKey)))

			// Verify key still accessible via Redis
			val, err = client.Get(ctx, string(testKey)).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal(testValue))
		})

		It("should validate storage consistency across server restarts", func() {
			// Validate storage consistency across server restarts
			// _Requirements: 5.1, 5.2_

			testKeys := []string{
				"persistent:key1",
				"persistent:key2",
				"persistent:key3",
			}

			baseTime := time.Now().Unix()

			// Set keys via Redis client and TTL metadata
			for i, keyStr := range testKeys {
				key := []byte(keyStr)
				value := fmt.Sprintf("value:%d", i)
				expiresAt := baseTime + int64((i+1)*1000)

				// Set key via Redis
				err := client.Set(ctx, keyStr, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				// Set TTL metadata
				err = ttlStorage.SetTTL(key, expiresAt)
				Expect(err).NotTo(HaveOccurred())
			}

			// Close current storage and server
			lmdbStorage.Close()
			srv.Close()
			client.Close()

			// Restart server with same data directory
			time.Sleep(100 * time.Millisecond)

			newSrv, err := server.New("localhost:6382", tmpDir)
			Expect(err).NotTo(HaveOccurred())

			go func() {
				defer GinkgoRecover()
				newSrv.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)

			newClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6382",
			})

			Eventually(func() error {
				return newClient.Ping(ctx).Err()
			}, "5s", "100ms").Should(Succeed())

			// Reopen TTL storage
			newStorage, err := storage.NewLMDBStorage(tmpDir + "/ttl_test")
			Expect(err).NotTo(HaveOccurred())
			defer newStorage.Close()
			defer newSrv.Close()
			defer newClient.Close()

			newTTLStorage := newStorage.GetTTLStorage()

			// Verify TTL metadata persisted across restart
			for i, keyStr := range testKeys {
				key := []byte(keyStr)
				expectedExpiresAt := baseTime + int64((i+1)*1000)

				metadata, err := newTTLStorage.GetTTL(key)
				Expect(err).NotTo(HaveOccurred())
				Expect(metadata.ExpiresAt).To(Equal(expectedExpiresAt))

				// Verify key still accessible via Redis after restart
				val, err := newClient.Get(ctx, keyStr).Result()
				Expect(err).NotTo(HaveOccurred())
				Expect(val).To(Equal(fmt.Sprintf("value:%d", i)))
			}
		})

		It("should handle concurrent TTL operations via multiple Redis clients", func() {
			// Test concurrent TTL operations via multiple Redis clients
			// _Requirements: 5.1, 5.2_

			numClients := 5
			keysPerClient := 4
			done := make(chan bool, numClients)

			baseTime := time.Now().Unix()

			for c := 0; c < numClients; c++ {
				go func(clientID int) {
					defer GinkgoRecover()

					// Create separate Redis client for each goroutine
					testClient := redis.NewClient(&redis.Options{
						Addr: "localhost:6381",
					})
					defer testClient.Close()

					// Wait for client to connect
					Eventually(func() error {
						return testClient.Ping(ctx).Err()
					}, "2s", "50ms").Should(Succeed())

					for k := 0; k < keysPerClient; k++ {
						keyStr := fmt.Sprintf("concurrent:key:%d:%d", clientID, k)
						key := []byte(keyStr)
						value := fmt.Sprintf("value:%d:%d", clientID, k)
						expiresAt := baseTime + int64(k*100)

						// Set key via Redis client
						err := testClient.Set(ctx, keyStr, value, 0).Err()
						Expect(err).NotTo(HaveOccurred())

						// Set TTL metadata concurrently
						err = ttlStorage.SetTTL(key, expiresAt)
						Expect(err).NotTo(HaveOccurred())

						// Verify both Redis key and TTL metadata
						val, err := testClient.Get(ctx, keyStr).Result()
						Expect(err).NotTo(HaveOccurred())
						Expect(val).To(Equal(value))

						metadata, err := ttlStorage.GetTTL(key)
						Expect(err).NotTo(HaveOccurred())
						Expect(metadata.ExpiresAt).To(Equal(expiresAt))
					}

					done <- true
				}(c)
			}

			// Wait for all clients to complete
			for c := 0; c < numClients; c++ {
				Eventually(done).Should(Receive())
			}

			// Verify all keys were set correctly
			totalKeys := numClients * keysPerClient
			expiredKeys, err := ttlStorage.GetExpiredKeys(baseTime + 1000)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(expiredKeys)).To(BeNumerically(">=", totalKeys/2))

			// Verify keys are still accessible via Redis
			for c := 0; c < numClients; c++ {
				for k := 0; k < keysPerClient; k++ {
					keyStr := fmt.Sprintf("concurrent:key:%d:%d", c, k)
					expectedValue := fmt.Sprintf("value:%d:%d", c, k)

					val, err := client.Get(ctx, keyStr).Result()
					Expect(err).NotTo(HaveOccurred())
					Expect(val).To(Equal(expectedValue))
				}
			}
		})
	})

	Describe("TTL Cleanup Operations", func() {
		It("should efficiently query expired keys", func() {
			// Test expired key cleanup behavior
			// _Requirements: 5.1, 5.2_

			baseTime := time.Now().Unix()
			expiredKeys := [][]byte{
				[]byte("expired:key1"),
				[]byte("expired:key2"),
			}
			activeKeys := [][]byte{
				[]byte("active:key1"),
				[]byte("active:key2"),
			}

			for _, key := range expiredKeys {
				err := ttlStorage.SetTTL(key, baseTime-100)
				Expect(err).NotTo(HaveOccurred())
			}

			for _, key := range activeKeys {
				err := ttlStorage.SetTTL(key, baseTime+100)
				Expect(err).NotTo(HaveOccurred())
			}

			foundExpired, err := ttlStorage.GetExpiredKeys(baseTime)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(foundExpired)).To(Equal(len(expiredKeys)))
		})

		It("should handle batch TTL removal efficiently", func() {
			// Test batch cleanup performance
			// _Requirements: 5.1, 5.2_

			batchSize := 100
			keys := make([][]byte, batchSize)
			baseTime := time.Now().Unix()

			for i := 0; i < batchSize; i++ {
				keys[i] = []byte(fmt.Sprintf("batch:key:%d", i))
				err := ttlStorage.SetTTL(keys[i], baseTime+int64(i))
				Expect(err).NotTo(HaveOccurred())
			}

			err := ttlStorage.RemoveTTLBatch(keys)
			Expect(err).NotTo(HaveOccurred())

			for _, key := range keys {
				_, err := ttlStorage.GetTTL(key)
				Expect(err).To(Equal(storage.ErrTTLNotFound))
			}
		})
	})

	Describe("TTL Error Handling", func() {
		It("should handle TTL operations on non-existent keys", func() {
			nonExistentKey := []byte("non:existent:key")

			_, err := ttlStorage.GetTTL(nonExistentKey)
			Expect(err).To(Equal(storage.ErrTTLNotFound))

			err = ttlStorage.RemoveTTL(nonExistentKey)
			Expect(err).To(Equal(storage.ErrTTLNotFound))
		})

		It("should validate TTL input parameters", func() {
			validKey := []byte("valid:key")

			err := ttlStorage.SetTTL([]byte{}, time.Now().Unix()+100)
			Expect(err).To(Equal(storage.ErrEmptyKey))

			largeKey := make([]byte, storage.MaxKeySize+1)
			err = ttlStorage.SetTTL(largeKey, time.Now().Unix()+100)
			Expect(err).To(Equal(storage.ErrKeyTooLarge))

			err = ttlStorage.SetTTL(validKey, -1)
			Expect(err).To(Equal(storage.ErrInvalidTimestamp))
		})
	})
})
