package server_test

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"sync"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/server"
)

var _ = Describe("TTL Performance Integration Tests", func() {
	var (
		srv    *server.Server
		client *redis.Client
		tmpDir string
		ctx    context.Context
	)

	BeforeEach(func() {
		os.Setenv("KEYP_TEST_MODE", "true")

		var err error
		tmpDir, err = os.MkdirTemp("", "keyp_ttl_performance_*")
		Expect(err).NotTo(HaveOccurred())

		ctx = context.Background()
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

	Describe("TTL Command Performance", func() {
		It("should handle high-volume EXPIRE operations efficiently", func() {
			keyCount := 10000
			keys := make([]string, keyCount)

			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("perf_expire_key_%d", i)
				keys[i] = key
				err := client.Set(ctx, key, "performance_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			start := time.Now()
			for _, key := range keys {
				result := client.Expire(ctx, key, time.Hour)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}
			duration := time.Since(start)

			avgPerOp := duration / time.Duration(keyCount)
			Expect(avgPerOp).To(BeNumerically("<", 1*time.Millisecond))

			GinkgoWriter.Printf("EXPIRE performance: %d operations in %v (avg: %v per operation)\n",
				keyCount, duration, avgPerOp)
		})

		It("should handle high-volume TTL queries efficiently", func() {
			keyCount := 10000
			keys := make([]string, keyCount)

			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("perf_ttl_key_%d", i)
				keys[i] = key
				err := client.Set(ctx, key, "performance_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
				err = client.Expire(ctx, key, time.Hour).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			start := time.Now()
			for _, key := range keys {
				result := client.TTL(ctx, key)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeNumerically(">", 0))
			}
			duration := time.Since(start)

			avgPerOp := duration / time.Duration(keyCount)
			Expect(avgPerOp).To(BeNumerically("<", 1*time.Millisecond))

			GinkgoWriter.Printf("TTL query performance: %d operations in %v (avg: %v per operation)\n",
				keyCount, duration, avgPerOp)
		})

		It("should handle concurrent TTL operations efficiently", func() {
			keyCount := 1000
			goroutineCount := 10
			var wg sync.WaitGroup

			start := time.Now()

			for g := 0; g < goroutineCount; g++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					defer GinkgoRecover()

					for i := 0; i < keyCount; i++ {
						key := fmt.Sprintf("concurrent_key_%d_%d", goroutineID, i)

						err := client.Set(ctx, key, "concurrent_value", 0).Err()
						Expect(err).NotTo(HaveOccurred())

						result := client.Expire(ctx, key, time.Hour)
						Expect(result.Err()).NotTo(HaveOccurred())
						Expect(result.Val()).To(BeTrue())

						ttlResult := client.TTL(ctx, key)
						Expect(ttlResult.Err()).NotTo(HaveOccurred())
						Expect(ttlResult.Val()).To(BeNumerically(">", 0))
					}
				}(g)
			}

			wg.Wait()
			duration := time.Since(start)

			totalOps := keyCount * goroutineCount * 3
			avgPerOp := duration / time.Duration(totalOps)

			GinkgoWriter.Printf("Concurrent TTL performance: %d operations across %d goroutines in %v (avg: %v per operation)\n",
				totalOps, goroutineCount, duration, avgPerOp)
		})
	})

	Describe("TTL Cleanup Performance", func() {
		It("should handle cleanup of large numbers of expired keys efficiently", func() {
			keyCount := 5000

			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("cleanup_perf_key_%d", i)
				err := client.Set(ctx, key, "cleanup_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())

				result := client.Expire(ctx, key, 1*time.Second)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			time.Sleep(2 * time.Second)

			start := time.Now()
			cleanupCount := 0
			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("cleanup_perf_key_%d", i)
				result := client.Get(ctx, key)
				if result.Err() == redis.Nil {
					cleanupCount++
				}
			}
			duration := time.Since(start)

			Expect(cleanupCount).To(BeNumerically(">", int(float64(keyCount)*0.8)))

			GinkgoWriter.Printf("Cleanup verification: %d keys checked in %v, %d were cleaned up\n",
				keyCount, duration, cleanupCount)
		})

		It("should maintain performance during mixed operations with cleanup", func() {
			activeKeyCount := 2000
			expiredKeyCount := 1000

			for i := 0; i < activeKeyCount; i++ {
				key := fmt.Sprintf("active_key_%d", i)
				err := client.Set(ctx, key, "active_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
				err = client.Expire(ctx, key, time.Hour).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			for i := 0; i < expiredKeyCount; i++ {
				key := fmt.Sprintf("expired_key_%d", i)
				err := client.Set(ctx, key, "expired_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
				err = client.Expire(ctx, key, 1*time.Second).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			time.Sleep(2 * time.Second)

			start := time.Now()
			for i := 0; i < activeKeyCount; i++ {
				key := fmt.Sprintf("active_key_%d", i)
				result := client.TTL(ctx, key)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeNumerically(">", 0))
			}
			duration := time.Since(start)

			avgPerOp := duration / time.Duration(activeKeyCount)
			Expect(avgPerOp).To(BeNumerically("<", 2*time.Millisecond))

			GinkgoWriter.Printf("Mixed operations performance: %d TTL queries in %v with %d expired keys present (avg: %v per operation)\n",
				activeKeyCount, duration, expiredKeyCount, avgPerOp)
		})
	})

	Describe("TTL Memory and Resource Usage", func() {
		It("should handle TTL operations without excessive memory growth", func() {
			keyCount := 20000
			batchSize := 1000

			for batch := 0; batch < keyCount/batchSize; batch++ {
				for i := 0; i < batchSize; i++ {
					key := fmt.Sprintf("memory_test_key_%d_%d", batch, i)
					err := client.Set(ctx, key, "memory_test_value", 0).Err()
					Expect(err).NotTo(HaveOccurred())

					result := client.Expire(ctx, key, time.Duration(batch+1)*time.Hour)
					Expect(result.Err()).NotTo(HaveOccurred())
					Expect(result.Val()).To(BeTrue())
				}

				if batch%5 == 0 {
					time.Sleep(10 * time.Millisecond)
				}
			}

			start := time.Now()
			validKeys := 0
			for batch := 0; batch < keyCount/batchSize; batch++ {
				for i := 0; i < batchSize; i++ {
					key := fmt.Sprintf("memory_test_key_%d_%d", batch, i)
					result := client.TTL(ctx, key)
					if result.Err() == nil && result.Val() > 0 {
						validKeys++
					}
				}
			}
			duration := time.Since(start)

			Expect(validKeys).To(BeNumerically(">=", int(float64(keyCount)*0.95)))

			GinkgoWriter.Printf("Memory test: %d keys processed, %d valid TTLs found in %v\n",
				keyCount, validKeys, duration)
		})
	})

	Describe("TTL Latency Measurements", func() {
		It("should measure TTL operation latencies under load", func() {
			keyCount := 1000
			measurements := make([]time.Duration, keyCount)

			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("latency_key_%d", i)
				err := client.Set(ctx, key, "latency_value", 0).Err()
				Expect(err).NotTo(HaveOccurred())
			}

			for i := 0; i < keyCount; i++ {
				key := fmt.Sprintf("latency_key_%d", i)

				start := time.Now()
				result := client.Expire(ctx, key, time.Hour)
				measurements[i] = time.Since(start)

				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(BeTrue())
			}

			var totalLatency time.Duration
			var maxLatency time.Duration
			minLatency := measurements[0]

			for _, latency := range measurements {
				totalLatency += latency
				if latency > maxLatency {
					maxLatency = latency
				}
				if latency < minLatency {
					minLatency = latency
				}
			}

			avgLatency := totalLatency / time.Duration(keyCount)

			Expect(avgLatency).To(BeNumerically("<", 5*time.Millisecond))
			Expect(maxLatency).To(BeNumerically("<", 50*time.Millisecond))

			GinkgoWriter.Printf("Latency measurements: avg=%v, min=%v, max=%v\n",
				avgLatency, minLatency, maxLatency)
		})
	})
})

func BenchmarkTTLRedisProtocol(b *testing.B) {
	os.Setenv("KEYP_TEST_MODE", "true")

	tmpDir, err := os.MkdirTemp("", "keyp_benchmark_redis_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srv, err := server.New("localhost:6382", tmpDir)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	go srv.ListenAndServe()
	defer srv.Close()

	time.Sleep(100 * time.Millisecond)

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6382",
	})
	defer client.Close()

	ctx := context.Background()
	err = client.Ping(ctx).Err()
	if err != nil {
		b.Fatalf("Failed to connect to server: %v", err)
	}

	b.Run("EXPIRE", func(b *testing.B) {
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_expire_%d", i)
			keys[i] = key
			client.Set(ctx, key, "bench_value", 0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Expire(ctx, keys[i], time.Hour)
		}
	})

	b.Run("TTL", func(b *testing.B) {
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_ttl_%d", i)
			keys[i] = key
			client.Set(ctx, key, "bench_value", 0)
			client.Expire(ctx, key, time.Hour)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.TTL(ctx, keys[i])
		}
	})

	b.Run("PTTL", func(b *testing.B) {
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_pttl_%d", i)
			keys[i] = key
			client.Set(ctx, key, "bench_value", 0)
			client.Expire(ctx, key, time.Hour)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.PTTL(ctx, keys[i])
		}
	})

	b.Run("EXPIREAT", func(b *testing.B) {
		keys := make([]string, b.N)
		timestamp := time.Now().Add(time.Hour)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_expireat_%d", i)
			keys[i] = key
			client.Set(ctx, key, "bench_value", 0)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.ExpireAt(ctx, keys[i], timestamp)
		}
	})

	b.Run("PERSIST", func(b *testing.B) {
		keys := make([]string, b.N)
		for i := 0; i < b.N; i++ {
			key := fmt.Sprintf("bench_persist_%d", i)
			keys[i] = key
			client.Set(ctx, key, "bench_value", 0)
			client.Expire(ctx, key, time.Hour)
		}

		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			client.Persist(ctx, keys[i])
		}
	})
}

func BenchmarkTTLConcurrentRedis(b *testing.B) {
	os.Setenv("KEYP_TEST_MODE", "true")

	tmpDir, err := os.MkdirTemp("", "keyp_benchmark_concurrent_redis_*")
	if err != nil {
		b.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	srv, err := server.New("localhost:6383", tmpDir)
	if err != nil {
		b.Fatalf("Failed to create server: %v", err)
	}

	go srv.ListenAndServe()
	defer srv.Close()

	time.Sleep(100 * time.Millisecond)

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6383",
	})
	defer client.Close()

	ctx := context.Background()

	b.Run("ConcurrentTTLOperations", func(b *testing.B) {
		b.RunParallel(func(pb *testing.PB) {
			i := 0
			for pb.Next() {
				key := "concurrent_key_" + strconv.Itoa(i)
				client.Set(ctx, key, "concurrent_value", 0)
				client.Expire(ctx, key, time.Hour)
				client.TTL(ctx, key)
				i++
			}
		})
	})
}
