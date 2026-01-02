package service_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"
	"github.com/tidwall/redcon"

	"github.com/luiz-simples/keyp.git/internal/service"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Handler Integration Tests", func() {
	var (
		handler     *service.Handler
		redisClient *redis.Client
		server      *redcon.Server
		ctx         context.Context
		storageImpl *storage.Client
		testPort    string
		testDir     string
	)

	BeforeEach(func() {
		ctx = context.Background()

		testDir = createUniqueTestDir("integration")

		var err error
		storageImpl, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		handler = service.NewHandler(storageImpl)

		listener, err := net.Listen("tcp", ":0")
		Expect(err).NotTo(HaveOccurred())
		testPort = fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
		listener.Close()

		server = redcon.NewServer(":"+testPort, func(conn redcon.Conn, cmd redcon.Command) {
			args := make([][]byte, len(cmd.Args))
			copy(args, cmd.Args)

			results := handler.Apply(ctx, args)

			if len(results) == 0 {
				conn.WriteError("ERR no results")
				return
			}

			result := results[0]
			if result.Error != nil {
				conn.WriteError(result.Error.Error())
				return
			}

			if result.Response == nil {
				conn.WriteNull()
				return
			}

			cmdName := strings.ToUpper(string(cmd.Args[0]))
			if len(cmd.Args) > 0 && (cmdName == "DEL" || cmdName == "DELETE") && len(result.Response) == 4 {
				count := binary.LittleEndian.Uint32(result.Response)
				conn.WriteInt64(int64(count))
				return
			}

			conn.WriteBulk(result.Response)
		}, nil, nil)

		go func() {
			server.ListenAndServe()
		}()

		time.Sleep(100 * time.Millisecond)

		redisClient = redis.NewClient(&redis.Options{
			Addr: "localhost:" + testPort,
		})

		Eventually(func() error {
			return redisClient.Ping(ctx).Err()
		}, "5s", "100ms").Should(Succeed())
	})

	AfterEach(func() {
		if redisClient != nil {
			redisClient.Close()
		}
		if server != nil {
			server.Close()
		}
		if storageImpl != nil {
			storageImpl.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("Basic Operations", func() {
		It("should handle PING command", func() {
			result := redisClient.Ping(ctx)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal("PONG"))
		})

		It("should handle PING with message", func() {
			message := "hello world"
			result := redisClient.Do(ctx, "PING", message)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal(message))
		})

		It("should handle SET and GET operations", func() {
			key := "integration:test:key"
			value := "integration test value"

			setResult := redisClient.Set(ctx, key, value, 0)
			Expect(setResult.Err()).NotTo(HaveOccurred())

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(value))
		})

		It("should handle GET for non-existent key", func() {
			result := redisClient.Get(ctx, "non:existent:key")
			Expect(result.Err()).To(Equal(redis.Nil))
		})

		It("should handle DEL operation", func() {
			key1 := "integration:del:key1"
			key2 := "integration:del:key2"
			value := "test value"

			redisClient.Set(ctx, key1, value, 0)
			redisClient.Set(ctx, key2, value, 0)

			delResult := redisClient.Del(ctx, key1, key2)
			Expect(delResult.Err()).NotTo(HaveOccurred())
			Expect(delResult.Val()).To(Equal(int64(2)))

			getResult1 := redisClient.Get(ctx, key1)
			Expect(getResult1.Err()).To(Equal(redis.Nil))

			getResult2 := redisClient.Get(ctx, key2)
			Expect(getResult2.Err()).To(Equal(redis.Nil))
		})
	})

	Describe("Complex Scenarios", func() {
		It("should handle multiple operations in sequence", func() {
			baseKey := "integration:sequence"

			for i := range 10 {
				key := fmt.Sprintf("%s:%d", baseKey, i)
				value := fmt.Sprintf("value_%d", i)

				setResult := redisClient.Set(ctx, key, value, 0)
				Expect(setResult.Err()).NotTo(HaveOccurred())
			}

			for i := range 10 {
				key := fmt.Sprintf("%s:%d", baseKey, i)
				expectedValue := fmt.Sprintf("value_%d", i)

				getResult := redisClient.Get(ctx, key)
				Expect(getResult.Err()).NotTo(HaveOccurred())
				Expect(getResult.Val()).To(Equal(expectedValue))
			}
		})

		It("should handle concurrent operations", func() {
			const numGoroutines = 10
			const operationsPerGoroutine = 100

			done := make(chan bool, numGoroutines)

			for g := range numGoroutines {
				go func(goroutineID int) {
					defer GinkgoRecover()

					for i := range operationsPerGoroutine {
						key := fmt.Sprintf("concurrent:%d:%d", goroutineID, i)
						value := fmt.Sprintf("value_%d_%d", goroutineID, i)

						setResult := redisClient.Set(ctx, key, value, 0)
						Expect(setResult.Err()).NotTo(HaveOccurred())

						getResult := redisClient.Get(ctx, key)
						Expect(getResult.Err()).NotTo(HaveOccurred())
						Expect(getResult.Val()).To(Equal(value))
					}

					done <- true
				}(g)
			}

			for range numGoroutines {
				Eventually(done).Should(Receive())
			}
		})

		It("should handle large values", func() {
			key := "integration:large:value"
			largeValue := make([]byte, 1024*1024)
			for i := range largeValue {
				largeValue[i] = byte(i % 256)
			}

			setResult := redisClient.Set(ctx, key, largeValue, 0)
			Expect(setResult.Err()).NotTo(HaveOccurred())

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect([]byte(getResult.Val())).To(Equal(largeValue))
		})
	})

	Describe("Error Handling", func() {
		It("should handle unknown commands", func() {
			result := redisClient.Do(ctx, "UNKNOWN_COMMAND")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("unknown command"))
		})

		It("should handle invalid argument counts", func() {
			result := redisClient.Do(ctx, "SET", "key")
			Expect(result.Err()).To(HaveOccurred())
			Expect(result.Err().Error()).To(ContainSubstring("wrong number of arguments"))
		})
	})
})
