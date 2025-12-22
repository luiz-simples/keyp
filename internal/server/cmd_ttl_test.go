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

var _ = Describe("TTL Commands Integration Tests", func() {
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
		tmpDir, err = os.MkdirTemp("", "ttl-cmd-test-*")
		Expect(err).NotTo(HaveOccurred())

		srv, err = server.New("localhost:6390", tmpDir)
		Expect(err).NotTo(HaveOccurred())

		go func() {
			defer GinkgoRecover()
			srv.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		client = redis.NewClient(&redis.Options{
			Addr: "localhost:6390",
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

	Describe("EXPIRE command", func() {
		It("should set TTL for existing keys", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", 3600*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 3500))
			Expect(ttl.Seconds()).To(BeNumerically("<=", 3600))
		})

		It("should return 0 for non-existent keys", func() {
			result, err := client.Expire(ctx, "nonexistent", 3600*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})

		It("should handle negative TTL", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", -100*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})

		It("should update existing TTL", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", 1000*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			result, err = client.Expire(ctx, "mykey", 2000*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 1900))
		})
	})

	Describe("EXPIREAT command", func() {
		It("should set absolute expiration time", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			futureTime := time.Now().Add(7200 * time.Second)
			result, err := client.ExpireAt(ctx, "mykey", futureTime).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 7100))
			Expect(ttl.Seconds()).To(BeNumerically("<=", 7200))
		})

		It("should return 0 for past timestamps", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			pastTime := time.Now().Add(-100 * time.Second)
			result, err := client.ExpireAt(ctx, "mykey", pastTime).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Describe("TTL command", func() {
		It("should return TTL in seconds", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", 1800*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 1700))
			Expect(ttl.Seconds()).To(BeNumerically("<=", 1800))
		})

		It("should return -1 for persistent keys", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(time.Duration(-1)))
		})

		It("should return -2 for non-existent keys", func() {
			ttl, err := client.TTL(ctx, "nonexistent").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(time.Duration(-2)))
		})
	})

	Describe("PTTL command", func() {
		It("should return TTL in milliseconds", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", 1800*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			pttl, err := client.PTTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl.Milliseconds()).To(BeNumerically(">", 1700000))
			Expect(pttl.Milliseconds()).To(BeNumerically("<=", 1800000))
		})

		It("should return -1 for persistent keys", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			pttl, err := client.PTTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl).To(Equal(time.Duration(-1)))
		})

		It("should return -2 for non-existent keys", func() {
			pttl, err := client.PTTL(ctx, "nonexistent").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl).To(Equal(time.Duration(-2)))
		})
	})

	Describe("PERSIST command", func() {
		It("should remove TTL from keys", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Expire(ctx, "mykey", 3600*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err := client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 0))

			persistResult, err := client.Persist(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(persistResult).To(BeTrue())

			ttl, err = client.TTL(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(time.Duration(-1)))
		})

		It("should return 0 for already persistent keys", func() {
			err := client.Set(ctx, "mykey", "myvalue", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			result, err := client.Persist(ctx, "mykey").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})

		It("should return 0 for non-existent keys", func() {
			result, err := client.Persist(ctx, "nonexistent").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeFalse())
		})
	})

	Describe("TTL Commands Integration", func() {
		It("should work together in a complete workflow", func() {
			err := client.Set(ctx, "workflow:key", "test:value", 0).Err()
			Expect(err).NotTo(HaveOccurred())

			ttl, err := client.TTL(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(time.Duration(-1)))

			result, err := client.Expire(ctx, "workflow:key", 5000*time.Second).Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(BeTrue())

			ttl, err = client.TTL(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl.Seconds()).To(BeNumerically(">", 4900))

			pttl, err := client.PTTL(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(pttl.Milliseconds()).To(BeNumerically(">", 4900000))

			persistResult, err := client.Persist(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(persistResult).To(BeTrue())

			ttl, err = client.TTL(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(ttl).To(Equal(time.Duration(-1)))

			val, err := client.Get(ctx, "workflow:key").Result()
			Expect(err).NotTo(HaveOccurred())
			Expect(val).To(Equal("test:value"))
		})
	})
})
