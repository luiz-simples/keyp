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

var _ = Describe("Command Handlers Edge Cases", func() {
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
		tmpDir, err = os.MkdirTemp("", "cmd-handlers-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6391", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6391",
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

	Describe("DEL command edge cases", func() {
		It("should handle DEL with wrong number of arguments", func() {
			result := client.Do(ctx, "DEL")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle DEL with multiple non-existent keys", func() {
			result := client.Del(ctx, "nonexistent1", "nonexistent2", "nonexistent3")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})

		It("should handle DEL with mix of existing and non-existent keys", func() {
			err := client.Set(ctx, "key1", "value1", 0).Err()
			Expect(err).NotTo(HaveOccurred())
			err = client.Set(ctx, "key3", "value3", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Del(ctx, "key1", "nonexistent", "key3", "another_nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(2))) // Only 2 keys existed
		})
	})

	Describe("GET command edge cases", func() {
		It("should handle GET with wrong number of arguments", func() {
			result := client.Do(ctx, "GET")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			result = client.Do(ctx, "GET", "key1", "key2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should return nil for non-existent key", func() {
			result := client.Get(ctx, "nonexistent")
			Expect(result.Err()).To(Equal(redis.Nil))
		})
	})

	Describe("SET command edge cases", func() {
		It("should handle SET with wrong number of arguments", func() {
			result := client.Do(ctx, "SET")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			result = client.Do(ctx, "SET", "key")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			result = client.Do(ctx, "SET", "key", "value", "extra")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle SET with empty key", func() {
			result := client.Do(ctx, "SET", "", "value")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("empty key"))
		})

		It("should handle SET with very large key", func() {
			largeKey := string(make([]byte, 1024))
			result := client.Do(ctx, "SET", largeKey, "value")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("key too large"))
		})
	})

	Describe("Server connection edge cases", func() {
		It("should handle unknown commands", func() {
			result := client.Do(ctx, "UNKNOWN_COMMAND", "arg1", "arg2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("unknown command"))
		})

		It("should handle PING with message", func() {
			result := client.Do(ctx, "PING", "test_message")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal("test_message"))
		})

		It("should handle PING without message", func() {
			result := client.Ping(ctx)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal("PONG"))
		})
	})

	Describe("TTL command error cases", func() {
		It("should handle TTL commands with storage errors", func() {
			result := client.Do(ctx, "TTL", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-2)))

			result = client.Do(ctx, "PTTL", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-2)))

			result = client.Do(ctx, "PERSIST", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})

		It("should handle TTL/PTTL commands with wrong arguments", func() {
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

			result = client.Do(ctx, "PERSIST")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			result = client.Do(ctx, "PERSIST", "key1", "key2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle EXPIRE/EXPIREAT with edge case values", func() {
			err := client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Do(ctx, "EXPIRE", "testkey", "3600") // 1 hour
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(1)))

			err = client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			pastTimestamp := time.Now().Unix() - 3600 // 1 hour ago
			result = client.Do(ctx, "EXPIREAT", "testkey", pastTimestamp)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0))) // Should fail

			futureTimestamp := time.Now().Unix() + 3600 // 1 hour from now
			result = client.Do(ctx, "EXPIREAT", "testkey", futureTimestamp)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(1))) // Should succeed
		})

		It("should handle EXPIRE/EXPIREAT with invalid arguments", func() {
			err := client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Do(ctx, "EXPIRE", "testkey", "invalid")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("value is not an integer"))

			result = client.Do(ctx, "EXPIREAT", "testkey", "invalid")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("value is not an integer"))

			result = client.Do(ctx, "EXPIRE")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			result = client.Do(ctx, "EXPIREAT")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle TTL operations on keys without TTL", func() {
			err := client.Set(ctx, "persistent_key", "value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Do(ctx, "TTL", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-1)))

			result = client.Do(ctx, "PTTL", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-1)))

			result = client.Do(ctx, "PERSIST", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})
	})
})
