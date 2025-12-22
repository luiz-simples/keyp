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

var _ = Describe("Expiration Integration Tests", func() {
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
		tmpDir, err = os.MkdirTemp("", "keyp-expiration-test-*")
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

	Describe("Automatic Expiration via Redis Protocol", func() {
		It("should expire keys automatically on GET operations", func() {
			key := "auto_expire_get"
			value := "expire_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			result := client.Expire(ctx, key, 1*time.Second)
			Expect(result.Err()).ToNot(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			retrievedValue := client.Get(ctx, key)
			Expect(retrievedValue.Err()).ToNot(HaveOccurred())
			Expect(retrievedValue.Val()).To(Equal(value))

			time.Sleep(2 * time.Second)

			retrievedValue = client.Get(ctx, key)
			Expect(retrievedValue.Err()).To(Equal(redis.Nil))
		})

		It("should handle TTL queries on expired keys", func() {
			key := "ttl_expired_key"
			value := "ttl_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			result := client.Expire(ctx, key, 1*time.Second)
			Expect(result.Err()).ToNot(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			ttl := client.TTL(ctx, key)
			Expect(ttl.Err()).ToNot(HaveOccurred())
			Expect(ttl.Val()).To(BeNumerically(">", 0))

			time.Sleep(2 * time.Second)

			ttl = client.TTL(ctx, key)
			Expect(ttl.Err()).ToNot(HaveOccurred())
			Expect(ttl.Val()).To(Equal(-2 * time.Nanosecond))
		})

		It("should handle PTTL queries on expired keys", func() {
			key := "pttl_expired_key"
			value := "pttl_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			result := client.Expire(ctx, key, 1*time.Second)
			Expect(result.Err()).ToNot(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			pttl := client.PTTL(ctx, key)
			Expect(pttl.Err()).ToNot(HaveOccurred())
			Expect(pttl.Val()).To(BeNumerically(">", 0))

			time.Sleep(2 * time.Second)

			pttl = client.PTTL(ctx, key)
			Expect(pttl.Err()).ToNot(HaveOccurred())
			Expect(pttl.Val()).To(Equal(-2 * time.Nanosecond))
		})

		It("should cleanup expired keys with DEL operations", func() {
			keys := []string{"del_exp_1", "del_exp_2", "del_exp_3"}
			value := "del_value"

			for _, key := range keys {
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).ToNot(HaveOccurred())

				result := client.Expire(ctx, key, 1*time.Second)
				Expect(result.Err()).ToNot(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			time.Sleep(2 * time.Second)

			for _, key := range keys {
				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).To(Equal(redis.Nil))
			}

			deletedCount := client.Del(ctx, keys...)
			Expect(deletedCount.Err()).ToNot(HaveOccurred())
			Expect(deletedCount.Val()).To(Equal(int64(0)))
		})

		It("should handle mixed operations with expired and persistent keys", func() {
			expiredKey := "mixed_expired"
			persistentKey := "mixed_persistent"
			value := "mixed_value"

			err := client.Set(ctx, expiredKey, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			err = client.Set(ctx, persistentKey, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			result := client.Expire(ctx, expiredKey, 1*time.Second)
			Expect(result.Err()).ToNot(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			time.Sleep(2 * time.Second)

			expiredValue := client.Get(ctx, expiredKey)
			Expect(expiredValue.Err()).To(Equal(redis.Nil))

			persistentValue := client.Get(ctx, persistentKey)
			Expect(persistentValue.Err()).ToNot(HaveOccurred())
			Expect(persistentValue.Val()).To(Equal(value))

			deletedCount := client.Del(ctx, expiredKey, persistentKey)
			Expect(deletedCount.Err()).ToNot(HaveOccurred())
			Expect(deletedCount.Val()).To(Equal(int64(1)))
		})

		It("should handle concurrent expiration operations", func() {
			baseKey := "concurrent_exp_"
			value := "concurrent_value"
			keyCount := 10

			for i := 0; i < keyCount; i++ {
				key := baseKey + string(rune('0'+i))
				err := client.Set(ctx, key, value, 0).Err()
				Expect(err).ToNot(HaveOccurred())

				result := client.Expire(ctx, key, 1*time.Second)
				Expect(result.Err()).ToNot(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			time.Sleep(2 * time.Second)

			for i := 0; i < keyCount; i++ {
				key := baseKey + string(rune('0'+i))
				retrievedValue := client.Get(ctx, key)
				Expect(retrievedValue.Err()).To(Equal(redis.Nil))
			}
		})

		It("should maintain expiration behavior with PERSIST operations", func() {
			key := "persist_then_expire"
			value := "persist_value"

			err := client.Set(ctx, key, value, 0).Err()
			Expect(err).ToNot(HaveOccurred())

			result := client.Expire(ctx, key, 10*time.Second)
			Expect(result.Err()).ToNot(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			persistResult := client.Persist(ctx, key)
			Expect(persistResult.Err()).ToNot(HaveOccurred())
			Expect(persistResult.Val()).To(BeTrue())

			newExpireResult := client.Expire(ctx, key, 1*time.Second)
			Expect(newExpireResult.Err()).ToNot(HaveOccurred())
			Expect(newExpireResult.Val()).To(BeTrue())

			time.Sleep(2 * time.Second)

			retrievedValue := client.Get(ctx, key)
			Expect(retrievedValue.Err()).To(Equal(redis.Nil))
		})
	})
})
