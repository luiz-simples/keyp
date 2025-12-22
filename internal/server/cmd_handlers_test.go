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

		// Set test mode to disable logging during tests
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
			// Test DEL with no arguments
			result := client.Do(ctx, "DEL")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle DEL with multiple non-existent keys", func() {
			// Test DEL with multiple non-existent keys
			result := client.Del(ctx, "nonexistent1", "nonexistent2", "nonexistent3")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})

		It("should handle DEL with mix of existing and non-existent keys", func() {
			// Set some keys
			err := client.Set(ctx, "key1", "value1", 0).Err()
			Expect(err).NotTo(HaveOccurred())
			err = client.Set(ctx, "key3", "value3", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Delete mix of existing and non-existent keys
			result := client.Del(ctx, "key1", "nonexistent", "key3", "another_nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(2))) // Only 2 keys existed
		})
	})

	Describe("GET command edge cases", func() {
		It("should handle GET with wrong number of arguments", func() {
			// Test GET with no arguments
			result := client.Do(ctx, "GET")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test GET with too many arguments
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
			// Test SET with no arguments
			result := client.Do(ctx, "SET")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test SET with only key (no value)
			result = client.Do(ctx, "SET", "key")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test SET with too many arguments
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
			// Create a key larger than MaxKeySize (assuming it's around 512 bytes)
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
			// These tests cover error paths in TTL commands
			// Test TTL on non-existent key
			result := client.Do(ctx, "TTL", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-2)))

			// Test PTTL on non-existent key
			result = client.Do(ctx, "PTTL", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-2)))

			// Test PERSIST on non-existent key
			result = client.Do(ctx, "PERSIST", "nonexistent")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})

		It("should handle TTL/PTTL commands with wrong arguments", func() {
			// Test TTL with no arguments
			result := client.Do(ctx, "TTL")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test TTL with too many arguments
			result = client.Do(ctx, "TTL", "key1", "key2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test PTTL with no arguments
			result = client.Do(ctx, "PTTL")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test PTTL with too many arguments
			result = client.Do(ctx, "PTTL", "key1", "key2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test PERSIST with no arguments
			result = client.Do(ctx, "PERSIST")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test PERSIST with too many arguments
			result = client.Do(ctx, "PERSIST", "key1", "key2")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle EXPIRE/EXPIREAT with edge case values", func() {
			// Set a key first
			err := client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Test EXPIRE with reasonable value (should work)
			result := client.Do(ctx, "EXPIRE", "testkey", "3600") // 1 hour
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(1)))

			// Reset key
			err = client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Test EXPIREAT with timestamp in the past (should fail)
			pastTimestamp := time.Now().Unix() - 3600 // 1 hour ago
			result = client.Do(ctx, "EXPIREAT", "testkey", pastTimestamp)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0))) // Should fail

			// Test EXPIREAT with future timestamp (should succeed)
			futureTimestamp := time.Now().Unix() + 3600 // 1 hour from now
			result = client.Do(ctx, "EXPIREAT", "testkey", futureTimestamp)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(1))) // Should succeed
		})

		It("should handle EXPIRE/EXPIREAT with invalid arguments", func() {
			// Set a key first
			err := client.Set(ctx, "testkey", "testvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Test EXPIRE with non-numeric value
			result := client.Do(ctx, "EXPIRE", "testkey", "invalid")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("value is not an integer"))

			// Test EXPIREAT with non-numeric value
			result = client.Do(ctx, "EXPIREAT", "testkey", "invalid")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("value is not an integer"))

			// Test EXPIRE with wrong number of arguments
			result = client.Do(ctx, "EXPIRE")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))

			// Test EXPIREAT with wrong number of arguments
			result = client.Do(ctx, "EXPIREAT")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})

		It("should handle TTL operations on keys without TTL", func() {
			// Set a key without TTL
			err := client.Set(ctx, "persistent_key", "value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			// Check TTL (should return -1 for keys without TTL)
			result := client.Do(ctx, "TTL", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-1)))

			// Check PTTL (should return -1 for keys without TTL)
			result = client.Do(ctx, "PTTL", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(-1)))

			// Try to persist (should return 0 since key has no TTL)
			result = client.Do(ctx, "PERSIST", "persistent_key")
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(int64(0)))
		})
	})
})
