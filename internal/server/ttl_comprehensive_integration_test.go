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

var _ = Describe("TTL Comprehensive Integration Tests", func() {
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
		tmpDir, err = os.MkdirTemp("", "ttl-comprehensive-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6382", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6382",
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

	Describe("TTL integration with existing SET/GET/DEL commands", func() {
		It("should integrate TTL with SET/GET workflow", func() {
			key := "ttl_set_get_key"
			value := "ttl_set_get_value"
			ttlSeconds := 30

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			retrievedValue := client.Get(ctx, key)
			Expect(retrievedValue.Err()).NotTo(HaveOccurred())
			Expect(retrievedValue.Val()).To(Equal(value))

			ttl := client.TTL(ctx, key)
			Expect(ttl.Err()).NotTo(HaveOccurred())
			Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(ttlSeconds-5)))
			Expect(ttl.Val().Seconds()).To(BeNumerically("<=", float64(ttlSeconds)))

			newValue := "updated_value"
			err = client.Set(ctx, key, newValue, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			updatedValue := client.Get(ctx, key)
			Expect(updatedValue.Err()).NotTo(HaveOccurred())
			Expect(updatedValue.Val()).To(Equal(newValue))

			ttlAfterSet := client.TTL(ctx, key)
			Expect(ttlAfterSet.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterSet.Val().Seconds()).To(BeNumerically(">", 0))
		})

		It("should integrate TTL with DEL operations", func() {
			keys := []string{"ttl_del_1", "ttl_del_2", "ttl_del_3"}
			value := "ttl_del_value"

			for _, key := range keys {
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 60*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			for _, key := range keys {
				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val().Seconds()).To(BeNumerically(">", 50))
			}

			deletedCount := client.Del(ctx, keys...)
			Expect(deletedCount.Err()).NotTo(HaveOccurred())
			Expect(deletedCount.Val()).To(Equal(int64(3)))

			for _, key := range keys {
				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).To(Equal(redis.Nil))

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val()).To(Equal(-2 * time.Nanosecond))
			}
		})

		It("should handle mixed TTL and persistent keys", func() {
			ttlKey := "mixed_ttl_key"
			persistentKey := "mixed_persistent_key"
			value := "mixed_value"

			err := client.Set(ctx, ttlKey, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())
			err = client.Set(ctx, persistentKey, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, ttlKey, 30*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			ttlValue := client.Get(ctx, ttlKey)
			Expect(ttlValue.Err()).NotTo(HaveOccurred())
			Expect(ttlValue.Val()).To(Equal(value))

			persistentValue := client.Get(ctx, persistentKey)
			Expect(persistentValue.Err()).NotTo(HaveOccurred())
			Expect(persistentValue.Val()).To(Equal(value))

			ttlResult := client.TTL(ctx, ttlKey)
			Expect(ttlResult.Err()).NotTo(HaveOccurred())
			Expect(ttlResult.Val().Seconds()).To(BeNumerically(">", 25))

			persistentTTL := client.TTL(ctx, persistentKey)
			Expect(persistentTTL.Err()).NotTo(HaveOccurred())
			Expect(persistentTTL.Val()).To(Equal(-1 * time.Nanosecond))

			deletedCount := client.Del(ctx, ttlKey, persistentKey)
			Expect(deletedCount.Err()).NotTo(HaveOccurred())
			Expect(deletedCount.Val()).To(Equal(int64(2)))
		})
	})

	Describe("TTL behavior during server restart", func() {
		It("should maintain TTL after server restart", func() {
			key := "restart_ttl_key"
			value := "restart_value"
			ttlSeconds := 120

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, key, time.Duration(ttlSeconds)*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			ttlBefore := client.TTL(ctx, key)
			Expect(ttlBefore.Err()).NotTo(HaveOccurred())
			Expect(ttlBefore.Val().Seconds()).To(BeNumerically(">", float64(ttlSeconds-10)))

			client.Close()
			srv.Close()

			time.Sleep(100 * time.Millisecond)

			newSrv, err := server.New("localhost:6382", tmpDir)
			Expect(err).NotTo(HaveOccurred())
			defer newSrv.Close()

			go func() {
				defer GinkgoRecover()
				newSrv.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)

			newClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6382",
			})
			defer newClient.Close()

			Eventually(func() error {
				return newClient.Ping(ctx).Err()
			}, "5s", "100ms").Should(Succeed())

			retrievedValue := newClient.Get(ctx, key)
			Expect(retrievedValue.Err()).NotTo(HaveOccurred())
			Expect(retrievedValue.Val()).To(Equal(value))

			ttlAfter := newClient.TTL(ctx, key)
			Expect(ttlAfter.Err()).NotTo(HaveOccurred())
			Expect(ttlAfter.Val().Seconds()).To(BeNumerically(">", float64(ttlSeconds-20)))
			Expect(ttlAfter.Val().Seconds()).To(BeNumerically("<=", ttlBefore.Val().Seconds()))

			srv = newSrv
			client = newClient
		})

		It("should cleanup expired keys during startup", func() {
			expiredKey := "startup_expired_key"
			activeKey := "startup_active_key"
			value := "startup_value"

			err := client.Set(ctx, expiredKey, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())
			err = client.Set(ctx, activeKey, value, 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, expiredKey, 1*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			result = client.Expire(ctx, activeKey, 300*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			time.Sleep(2 * time.Second)

			client.Close()
			srv.Close()

			time.Sleep(100 * time.Millisecond)

			newSrv, err := server.New("localhost:6382", tmpDir)
			Expect(err).NotTo(HaveOccurred())
			defer newSrv.Close()

			go func() {
				defer GinkgoRecover()
				newSrv.ListenAndServe()
			}()

			time.Sleep(100 * time.Millisecond)

			newClient := redis.NewClient(&redis.Options{
				Addr: "localhost:6382",
			})
			defer newClient.Close()

			Eventually(func() error {
				return newClient.Ping(ctx).Err()
			}, "5s", "100ms").Should(Succeed())

			expiredValue := newClient.Get(ctx, expiredKey)
			Expect(expiredValue.Err()).To(Equal(redis.Nil))

			activeValue := newClient.Get(ctx, activeKey)
			Expect(activeValue.Err()).NotTo(HaveOccurred())
			Expect(activeValue.Val()).To(Equal(value))

			srv = newSrv
			client = newClient
		})
	})

	Describe("Concurrent TTL operations from multiple Redis clients", func() {
		It("should handle concurrent TTL operations safely", func() {
			numClients := 3
			keysPerClient := 5
			clients := make([]*redis.Client, numClients)
			done := make(chan bool, numClients)

			for i := 0; i < numClients; i++ {
				clients[i] = redis.NewClient(&redis.Options{
					Addr: "localhost:6382",
				})
				defer clients[i].Close()

				Eventually(func() error {
					return clients[i].Ping(ctx).Err()
				}, "2s", "50ms").Should(Succeed())
			}

			for clientID := 0; clientID < numClients; clientID++ {
				go func(id int) {
					defer GinkgoRecover()
					testClient := clients[id]

					for keyID := 0; keyID < keysPerClient; keyID++ {
						key := "concurrent_" + string(rune('A'+id)) + "_" + string(rune('0'+keyID))
						value := "value_" + string(rune('A'+id)) + "_" + string(rune('0'+keyID))

						err := testClient.Set(ctx, key, value, 0).Err()
						Expect(err).NotTo(HaveOccurred())

						result := testClient.Expire(ctx, key, time.Duration(60+keyID)*time.Second)
						Expect(result.Err()).NotTo(HaveOccurred())
						Expect(result.Val()).To(BeTrue())

						retrievedValue := testClient.Get(ctx, key)
						Expect(retrievedValue.Err()).NotTo(HaveOccurred())
						Expect(retrievedValue.Val()).To(Equal(value))

						ttl := testClient.TTL(ctx, key)
						Expect(ttl.Err()).NotTo(HaveOccurred())
						Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(55+keyID)))
					}

					done <- true
				}(clientID)
			}

			for i := 0; i < numClients; i++ {
				Eventually(done).Should(Receive())
			}

			for clientID := 0; clientID < numClients; clientID++ {
				for keyID := 0; keyID < keysPerClient; keyID++ {
					key := "concurrent_" + string(rune('A'+clientID)) + "_" + string(rune('0'+keyID))
					value := "value_" + string(rune('A'+clientID)) + "_" + string(rune('0'+keyID))

					retrievedValue := client.Get(ctx, key)
					Expect(retrievedValue.Err()).NotTo(HaveOccurred())
					Expect(retrievedValue.Val()).To(Equal(value))
				}
			}
		})
	})

	Describe("TTL metadata consistency validation", func() {
		It("should maintain TTL metadata consistency across operations", func() {
			baseKey := "consistency_key_"
			value := "consistency_value"
			keyCount := 10

			for i := 0; i < keyCount; i++ {
				key := baseKey + string(rune('0'+i))
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).NotTo(HaveOccurred())

				if i%2 == 0 {
					result := client.Expire(ctx, key, time.Duration(120+i)*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())
				}
			}

			for i := 0; i < keyCount; i++ {
				key := baseKey + string(rune('0'+i))

				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).NotTo(HaveOccurred())
				Expect(retrievedValue.Val()).To(Equal(value))

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())

				if i%2 == 0 {
					Expect(ttl.Val().Seconds()).To(BeNumerically(">", float64(115+i)))
				}
			if i%2 != 0 {
					Expect(ttl.Val()).To(Equal(-1 * time.Nanosecond))
				}
			}

			keysToDelete := []string{
				baseKey + "0",
				baseKey + "2",
				baseKey + "5",
				baseKey + "7",
			}

			deletedCount := client.Del(ctx, keysToDelete...)
			Expect(deletedCount.Err()).NotTo(HaveOccurred())
			Expect(deletedCount.Val()).To(Equal(int64(4)))

			for _, key := range keysToDelete {
				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).To(Equal(redis.Nil))

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val()).To(Equal(-2 * time.Nanosecond))
			}

			remainingKeys := []string{
				baseKey + "1",
				baseKey + "3",
				baseKey + "4",
				baseKey + "6",
				baseKey + "8",
				baseKey + "9",
			}

			for _, key := range remainingKeys {
				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).NotTo(HaveOccurred())
				Expect(retrievedValue.Val()).To(Equal(value))
			}
		})
	})
})
