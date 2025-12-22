package server

import (
	"context"
	"os"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
)

var _ = Describe("TTL Persistence Integration Tests", func() {
	var (
		server     *Server
		client     *redis.Client
		ctx        context.Context
		tempDir    string
		serverAddr = "localhost:6380"
	)

	BeforeEach(func() {
		os.Setenv("KEYP_TEST_MODE", "true")
		ctx = context.Background()

		var err error
		tempDir, err = os.MkdirTemp("", "keyp-ttl-persistence-test")
		Expect(err).NotTo(HaveOccurred())

		server, err = New(serverAddr, tempDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			serverErr := server.ListenAndServe()
			if HasError(serverErr) {
				GinkgoLogr.Error(serverErr, "Server failed to start")
			}
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: serverAddr,
		})

		err = client.Ping(ctx).Err()
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		if server != nil {
			server.Close()
		}
		os.RemoveAll(tempDir)
	})

	restartServer := func() {
		client.Close()
		server.Close()

		var err error
		server, err = New(serverAddr, tempDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			serverErr := server.ListenAndServe()
			if HasError(serverErr) {
				GinkgoLogr.Error(serverErr, "Server failed to restart")
			}
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: serverAddr,
		})

		err = client.Ping(ctx).Err()
		Expect(err).NotTo(HaveOccurred())
	}

	Describe("TTL Persistence Across Server Restarts", func() {
		It("should maintain TTL values after server restart", func() {
			err := client.Set(ctx, "persistent-key", "test-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, "persistent-key", 300*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			originalTTL := client.TTL(ctx, "persistent-key")
			Expect(originalTTL.Err()).NotTo(HaveOccurred())
			Expect(originalTTL.Val()).To(BeNumerically(">", 290*time.Second))

			restartServer()

			value := client.Get(ctx, "persistent-key")
			Expect(value.Err()).NotTo(HaveOccurred())
			Expect(value.Val()).To(Equal("test-value"))

			restoredTTL := client.TTL(ctx, "persistent-key")
			Expect(restoredTTL.Err()).NotTo(HaveOccurred())
			Expect(restoredTTL.Val()).To(BeNumerically(">", 285*time.Second))
			Expect(restoredTTL.Val()).To(BeNumerically("<=", originalTTL.Val()))
		})

		It("should maintain PTTL precision after server restart", func() {
			err := client.Set(ctx, "precision-key", "test-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, "precision-key", 120*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			originalPTTL := client.PTTL(ctx, "precision-key")
			Expect(originalPTTL.Err()).NotTo(HaveOccurred())
			Expect(originalPTTL.Val()).To(BeNumerically(">", 115000*time.Millisecond))

			restartServer()

			restoredPTTL := client.PTTL(ctx, "precision-key")
			Expect(restoredPTTL.Err()).NotTo(HaveOccurred())
			Expect(restoredPTTL.Val()).To(BeNumerically(">", 110000*time.Millisecond))
			Expect(restoredPTTL.Val()).To(BeNumerically("<=", originalPTTL.Val()))
		})

		It("should handle EXPIREAT persistence correctly", func() {
			futureTimestamp := time.Now().Add(5 * time.Minute).Unix()

			err := client.Set(ctx, "expireat-key", "test-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.ExpireAt(ctx, "expireat-key", time.Unix(futureTimestamp, 0))
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			originalTTL := client.TTL(ctx, "expireat-key")
			Expect(originalTTL.Err()).NotTo(HaveOccurred())
			Expect(originalTTL.Val()).To(BeNumerically(">", 290*time.Second))

			restartServer()

			value := client.Get(ctx, "expireat-key")
			Expect(value.Err()).NotTo(HaveOccurred())
			Expect(value.Val()).To(Equal("test-value"))

			restoredTTL := client.TTL(ctx, "expireat-key")
			Expect(restoredTTL.Err()).NotTo(HaveOccurred())
			Expect(restoredTTL.Val()).To(BeNumerically(">", 285*time.Second))
		})
	})

	Describe("Expired Key Cleanup on Startup", func() {
		It("should remove expired keys during server startup", func() {
			err := client.Set(ctx, "short-lived-key", "test-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, "short-lived-key", 1*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			ttl := client.TTL(ctx, "short-lived-key")
			Expect(ttl.Err()).NotTo(HaveOccurred())
			Expect(ttl.Val()).To(BeNumerically(">", 0))

			time.Sleep(2 * time.Second)

			restartServer()

			value := client.Get(ctx, "short-lived-key")
			Expect(value.Err()).To(Equal(redis.Nil))

			ttlAfterRestart := client.TTL(ctx, "short-lived-key")
			Expect(ttlAfterRestart.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterRestart.Val()).To(Equal(-2 * time.Nanosecond))
		})

		It("should cleanup multiple expired keys on startup", func() {
			keys := []string{"expired-1", "expired-2", "expired-3"}

			for _, key := range keys {
				err := client.Set(ctx, key, "test-value", 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 1*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			time.Sleep(2 * time.Second)

			restartServer()

			for _, key := range keys {
				value := client.Get(ctx, key)
				Expect(value.Err()).To(Equal(redis.Nil))

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())
				Expect(ttl.Val()).To(Equal(-2 * time.Nanosecond))
			}
		})
	})

	Describe("Mixed TTL States Persistence", func() {
		It("should handle mix of persistent and TTL keys correctly", func() {
			err := client.Set(ctx, "persistent-key", "persistent-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			err = client.Set(ctx, "ttl-key", "ttl-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, "ttl-key", 180*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			persistentTTL := client.TTL(ctx, "persistent-key")
			Expect(persistentTTL.Err()).NotTo(HaveOccurred())
			Expect(persistentTTL.Val()).To(Equal(-1 * time.Nanosecond))

			ttlKeyTTL := client.TTL(ctx, "ttl-key")
			Expect(ttlKeyTTL.Err()).NotTo(HaveOccurred())
			Expect(ttlKeyTTL.Val()).To(BeNumerically(">", 175*time.Second))

			restartServer()

			persistentValue := client.Get(ctx, "persistent-key")
			Expect(persistentValue.Err()).NotTo(HaveOccurred())
			Expect(persistentValue.Val()).To(Equal("persistent-value"))

			ttlValue := client.Get(ctx, "ttl-key")
			Expect(ttlValue.Err()).NotTo(HaveOccurred())
			Expect(ttlValue.Val()).To(Equal("ttl-value"))

			persistentTTLAfter := client.TTL(ctx, "persistent-key")
			Expect(persistentTTLAfter.Err()).NotTo(HaveOccurred())
			Expect(persistentTTLAfter.Val()).To(Equal(-1 * time.Nanosecond))

			ttlKeyTTLAfter := client.TTL(ctx, "ttl-key")
			Expect(ttlKeyTTLAfter.Err()).NotTo(HaveOccurred())
			Expect(ttlKeyTTLAfter.Val()).To(BeNumerically(">", 170*time.Second))
		})
	})

	Describe("TTL Operations After Restart", func() {
		It("should allow TTL operations on restored keys", func() {
			err := client.Set(ctx, "modifiable-key", "test-value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result := client.Expire(ctx, "modifiable-key", 300*time.Second)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(BeTrue())

			restartServer()

			value := client.Get(ctx, "modifiable-key")
			Expect(value.Err()).NotTo(HaveOccurred())
			Expect(value.Val()).To(Equal("test-value"))

			persistResult := client.Persist(ctx, "modifiable-key")
			Expect(persistResult.Err()).NotTo(HaveOccurred())
			Expect(persistResult.Val()).To(BeTrue())

			ttlAfterPersist := client.TTL(ctx, "modifiable-key")
			Expect(ttlAfterPersist.Err()).NotTo(HaveOccurred())
			Expect(ttlAfterPersist.Val()).To(Equal(-1 * time.Nanosecond))

			newExpireResult := client.Expire(ctx, "modifiable-key", 60*time.Second)
			Expect(newExpireResult.Err()).NotTo(HaveOccurred())
			Expect(newExpireResult.Val()).To(BeTrue())

			newTTL := client.TTL(ctx, "modifiable-key")
			Expect(newTTL.Err()).NotTo(HaveOccurred())
			Expect(newTTL.Val()).To(BeNumerically(">", 55*time.Second))
		})
	})

	Describe("Concurrent TTL Operations During Restart", func() {
		It("should handle concurrent operations correctly", func() {
			keys := make([]string, 10)
			for i := 0; i < 10; i++ {
				keys[i] = "concurrent-key-" + strconv.Itoa(i)
				err := client.Set(ctx, keys[i], "value-"+strconv.Itoa(i), 0).Err()
				Expect(err).NotTo(HaveOccurred())

				if i%2 == 0 {
					result := client.Expire(ctx, keys[i], 240*time.Second)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())
				}
			}

			restartServer()

			for i, key := range keys {
				value := client.Get(ctx, key)
				Expect(value.Err()).NotTo(HaveOccurred())
				Expect(value.Val()).To(Equal("value-" + strconv.Itoa(i)))

				ttl := client.TTL(ctx, key)
				Expect(ttl.Err()).NotTo(HaveOccurred())

				if i%2 == 0 {
					Expect(ttl.Val()).To(BeNumerically(">", 230*time.Second))
				}
				if i%2 != 0 {
					Expect(ttl.Val()).To(Equal(-1 * time.Nanosecond))
				}
			}
		})
	})
})
