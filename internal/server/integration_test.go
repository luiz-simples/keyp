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
			_ = srv.ListenAndServe()
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
