package service_test

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/redis/go-redis/v9"

	"github.com/luiz-simples/keyp.git/internal/app"
	"github.com/luiz-simples/keyp.git/internal/service"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Redis Compatibility Integration Tests", func() {
	var (
		redisClient *redis.Client
		server      *app.Server
		ctx         context.Context
		storageImpl *storage.Client
		testPort    string
		testDir     string
		poolService *service.Pool
	)

	// createRedisClient cria um cliente Redis com configurações otimizadas para testes
	createRedisClient := func(addr string) *redis.Client {
		return redis.NewClient(&redis.Options{
			Addr:         addr,
			PoolSize:     1,               // Usar apenas 1 conexão para evitar problems de pool
			MinIdleConns: 0,               // Não manter conexões idle
			MaxRetries:   1,               // Reduzir tentativas de retry
			DialTimeout:  time.Second * 5, // Timeout para estabelecer conexão
			ReadTimeout:  time.Second * 3, // Timeout para leitura
			WriteTimeout: time.Second * 3, // Timeout para escrita
			PoolTimeout:  time.Second * 4, // Timeout para obter conexão do pool
		})
	}

	BeforeEach(func() {
		ctx = context.Background()
		testDir = createUniqueTestDir("redis-integration")

		var err error
		storageImpl, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())

		poolService = service.NewPool(storageImpl)
		server = app.NewServer(poolService)

		listener, err := net.Listen("tcp", ":0")
		Expect(err).NotTo(HaveOccurred())
		testPort = fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
		listener.Close()

		config := app.Config{
			Address: "localhost:" + testPort,
			DataDir: testDir,
		}

		ch := make(chan bool)
		go func(chBool chan bool) {
			defer GinkgoRecover()
			chBool <- true
			err := server.Start(config)
			if err != nil {
				fmt.Printf("Server error: %v\n", err)
			}
		}(ch)

		<-ch

		redisClient = createRedisClient("localhost:" + testPort)

		Eventually(func() error {
			return redisClient.Ping(ctx).Err()
		}, "10s", "100ms").Should(Succeed())
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

	Describe("String Operations", func() {
		It("should handle SET and GET commands", func() {
			key := "test:string:key"
			value := "test string value"

			result := redisClient.Set(ctx, key, value, 0)
			Expect(result.Err()).NotTo(HaveOccurred())
			Expect(result.Val()).To(Equal("OK"))

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(value))
		})

		It("should handle APPEND command", func() {
			key := "test:append:key"
			value1 := "Hello"
			value2 := " World"

			redisClient.Set(ctx, key, value1, 0)

			appendResult := redisClient.Append(ctx, key, value2)
			Expect(appendResult.Err()).NotTo(HaveOccurred())
			Expect(appendResult.Val()).To(Equal(int64(len(value1 + value2))))

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(value1 + value2))
		})

		It("should handle INCR command", func() {
			key := "test:incr:key"

			incrResult := redisClient.Incr(ctx, key)
			Expect(incrResult.Err()).NotTo(HaveOccurred())
			Expect(incrResult.Val()).To(Equal(int64(1)))

			incrResult = redisClient.Incr(ctx, key)
			Expect(incrResult.Err()).NotTo(HaveOccurred())
			Expect(incrResult.Val()).To(Equal(int64(2)))
		})

		It("should handle INCRBY command", func() {
			key := "test:incrby:key"
			increment := int64(5)

			incrByResult := redisClient.IncrBy(ctx, key, increment)
			Expect(incrByResult.Err()).NotTo(HaveOccurred())
			Expect(incrByResult.Val()).To(Equal(increment))

			incrByResult = redisClient.IncrBy(ctx, key, increment)
			Expect(incrByResult.Err()).NotTo(HaveOccurred())
			Expect(incrByResult.Val()).To(Equal(increment * 2))
		})

		It("should handle DECR command", func() {
			key := "test:decr:key"

			redisClient.Set(ctx, key, "10", 0)

			decrResult := redisClient.Decr(ctx, key)
			Expect(decrResult.Err()).NotTo(HaveOccurred())
			Expect(decrResult.Val()).To(Equal(int64(9)))

			decrResult = redisClient.Decr(ctx, key)
			Expect(decrResult.Err()).NotTo(HaveOccurred())
			Expect(decrResult.Val()).To(Equal(int64(8)))
		})

		It("should handle DECRBY command", func() {
			key := "test:decrby:key"
			decrement := int64(3)

			redisClient.Set(ctx, key, "10", 0)

			decrByResult := redisClient.DecrBy(ctx, key, decrement)
			Expect(decrByResult.Err()).NotTo(HaveOccurred())
			Expect(decrByResult.Val()).To(Equal(int64(7)))

			decrByResult = redisClient.DecrBy(ctx, key, decrement)
			Expect(decrByResult.Err()).NotTo(HaveOccurred())
			Expect(decrByResult.Val()).To(Equal(int64(4)))
		})
	})

	Describe("Key Operations", func() {
		It("should handle EXISTS command", func() {
			key := "test:exists:key"
			value := "test value"

			existsResult := redisClient.Exists(ctx, key)
			Expect(existsResult.Err()).NotTo(HaveOccurred())
			Expect(existsResult.Val()).To(Equal(int64(0)))

			redisClient.Set(ctx, key, value, 0)

			existsResult = redisClient.Exists(ctx, key)
			Expect(existsResult.Err()).NotTo(HaveOccurred())
			Expect(existsResult.Val()).To(Equal(int64(1)))
		})

		It("should handle DEL command", func() {
			key1 := "test:del:key1"
			key2 := "test:del:key2"
			value := "test value"

			redisClient.Set(ctx, key1, value, 0)
			redisClient.Set(ctx, key2, value, 0)

			delResult := redisClient.Del(ctx, key1, key2)
			Expect(delResult.Err()).NotTo(HaveOccurred())
			Expect(delResult.Val()).To(Equal(int64(2)))

			existsResult := redisClient.Exists(ctx, key1, key2)
			Expect(existsResult.Err()).NotTo(HaveOccurred())
			Expect(existsResult.Val()).To(Equal(int64(0)))
		})

		It("should handle EXPIRE and TTL commands", func() {
			key := "test:expire:key"
			value := "test value"
			expireSeconds := 2

			redisClient.Set(ctx, key, value, 0)

			expireResult := redisClient.Expire(ctx, key, time.Duration(expireSeconds)*time.Second)
			Expect(expireResult.Err()).NotTo(HaveOccurred())
			Expect(expireResult.Val()).To(BeTrue())

			ttlResult := redisClient.TTL(ctx, key)
			Expect(ttlResult.Err()).NotTo(HaveOccurred())
			Expect(ttlResult.Val().Seconds()).To(BeNumerically("<=", float64(expireSeconds)))
			Expect(ttlResult.Val().Seconds()).To(BeNumerically(">", 0))
		})

		It("should handle PERSIST command", func() {
			key := "test:persist:key"
			value := "test value"

			redisClient.Set(ctx, key, value, 0)
			redisClient.Expire(ctx, key, 10*time.Second)

			persistResult := redisClient.Persist(ctx, key)
			Expect(persistResult.Err()).NotTo(HaveOccurred())
			Expect(persistResult.Val()).To(BeTrue())

			ttlResult := redisClient.TTL(ctx, key)
			Expect(ttlResult.Err()).NotTo(HaveOccurred())
			Expect(ttlResult.Val()).To(Equal(time.Duration(-1)))
		})
	})

	Describe("List Operations", func() {
		It("should handle LPUSH and LLEN commands", func() {
			key := "test:list:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				lpushResult := redisClient.LPush(ctx, key, value)
				Expect(lpushResult.Err()).NotTo(HaveOccurred())
			}

			llenResult := redisClient.LLen(ctx, key)
			Expect(llenResult.Err()).NotTo(HaveOccurred())
			Expect(llenResult.Val()).To(Equal(int64(len(values))))
		})

		It("should handle RPUSH command", func() {
			key := "test:rpush:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				rpushResult := redisClient.RPush(ctx, key, value)
				Expect(rpushResult.Err()).NotTo(HaveOccurred())
			}

			llenResult := redisClient.LLen(ctx, key)
			Expect(llenResult.Err()).NotTo(HaveOccurred())
			Expect(llenResult.Val()).To(Equal(int64(len(values))))
		})

		It("should handle LRANGE command", func() {
			key := "test:lrange:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				redisClient.RPush(ctx, key, value)
			}

			lrangeResult := redisClient.LRange(ctx, key, 0, -1)
			Expect(lrangeResult.Err()).NotTo(HaveOccurred())
			Expect(lrangeResult.Val()).To(Equal(values))
		})

		It("should handle LINDEX command", func() {
			key := "test:lindex:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				redisClient.RPush(ctx, key, value)
			}

			lindexResult := redisClient.LIndex(ctx, key, 1)
			Expect(lindexResult.Err()).NotTo(HaveOccurred())
			Expect(lindexResult.Val()).To(Equal(values[1]))
		})

		It("should handle LSET command", func() {
			key := "test:lset:key"
			values := []string{"value1", "value2", "value3"}
			newValue := "new_value"

			for _, value := range values {
				redisClient.RPush(ctx, key, value)
			}

			lsetResult := redisClient.LSet(ctx, key, 1, newValue)
			Expect(lsetResult.Err()).NotTo(HaveOccurred())
			Expect(lsetResult.Val()).To(Equal("OK"))

			lindexResult := redisClient.LIndex(ctx, key, 1)
			Expect(lindexResult.Err()).NotTo(HaveOccurred())
			Expect(lindexResult.Val()).To(Equal(newValue))
		})

		It("should handle LPOP command", func() {
			key := "test:lpop:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				redisClient.LPush(ctx, key, value)
			}

			lpopResult := redisClient.LPop(ctx, key)
			Expect(lpopResult.Err()).NotTo(HaveOccurred())
			Expect(lpopResult.Val()).To(Equal(values[len(values)-1]))

			llenResult := redisClient.LLen(ctx, key)
			Expect(llenResult.Err()).NotTo(HaveOccurred())
			Expect(llenResult.Val()).To(Equal(int64(len(values) - 1)))
		})

		It("should handle RPOP command", func() {
			key := "test:rpop:key"
			values := []string{"value1", "value2", "value3"}

			for _, value := range values {
				redisClient.RPush(ctx, key, value)
			}

			rpopResult := redisClient.RPop(ctx, key)
			Expect(rpopResult.Err()).NotTo(HaveOccurred())
			Expect(rpopResult.Val()).To(Equal(values[len(values)-1]))

			llenResult := redisClient.LLen(ctx, key)
			Expect(llenResult.Err()).NotTo(HaveOccurred())
			Expect(llenResult.Val()).To(Equal(int64(len(values) - 1)))
		})
	})

	Describe("Set Operations", func() {
		It("should handle SADD and SMEMBERS commands", func() {
			key := "test:set:key"
			members := []string{"member1", "member2", "member3"}

			for _, member := range members {
				saddResult := redisClient.SAdd(ctx, key, member)
				Expect(saddResult.Err()).NotTo(HaveOccurred())
			}

			smembersResult := redisClient.SMembers(ctx, key)
			Expect(smembersResult.Err()).NotTo(HaveOccurred())
			Expect(smembersResult.Val()).To(ConsistOf(members))
		})

		It("should handle SISMEMBER command", func() {
			key := "test:sismember:key"
			member := "test_member"

			saddResult := redisClient.SAdd(ctx, key, member)
			Expect(saddResult.Err()).NotTo(HaveOccurred())

			sismemberResult := redisClient.SIsMember(ctx, key, member)
			Expect(sismemberResult.Err()).NotTo(HaveOccurred())
			Expect(sismemberResult.Val()).To(BeTrue())

			sismemberResult = redisClient.SIsMember(ctx, key, "non_existent")
			Expect(sismemberResult.Err()).NotTo(HaveOccurred())
			Expect(sismemberResult.Val()).To(BeFalse())
		})

		It("should handle SREM command", func() {
			key := "test:srem:key"
			members := []string{"member1", "member2", "member3"}

			for _, member := range members {
				redisClient.SAdd(ctx, key, member)
			}

			sremResult := redisClient.SRem(ctx, key, members[0])
			Expect(sremResult.Err()).NotTo(HaveOccurred())
			Expect(sremResult.Val()).To(Equal(int64(1)))

			sismemberResult := redisClient.SIsMember(ctx, key, members[0])
			Expect(sismemberResult.Err()).NotTo(HaveOccurred())
			Expect(sismemberResult.Val()).To(BeFalse())
		})
	})

	Describe("Sorted Set Operations", func() {
		It("should handle ZADD and ZRANGE commands", func() {
			key := "test:zset:key"
			members := []redis.Z{
				{Score: 1.0, Member: "member1"},
				{Score: 2.0, Member: "member2"},
				{Score: 3.0, Member: "member3"},
			}

			for _, member := range members {
				zaddResult := redisClient.ZAdd(ctx, key, member)
				Expect(zaddResult.Err()).NotTo(HaveOccurred())
			}

			zrangeResult := redisClient.ZRange(ctx, key, 0, -1)
			Expect(zrangeResult.Err()).NotTo(HaveOccurred())
			expectedMembers := []string{"member1", "member2", "member3"}
			Expect(zrangeResult.Val()).To(Equal(expectedMembers))
		})

		It("should handle ZCOUNT command", func() {
			key := "test:zcount:key"
			members := []redis.Z{
				{Score: 1.0, Member: "member1"},
				{Score: 2.0, Member: "member2"},
				{Score: 3.0, Member: "member3"},
				{Score: 4.0, Member: "member4"},
			}

			for _, member := range members {
				redisClient.ZAdd(ctx, key, member)
			}

			zcountResult := redisClient.ZCount(ctx, key, "2", "3")
			Expect(zcountResult.Err()).NotTo(HaveOccurred())
			Expect(zcountResult.Val()).To(Equal(int64(2)))
		})
	})

	Describe("Database Operations", func() {
		It("should handle PING command", func() {
			pingResult := redisClient.Ping(ctx)
			Expect(pingResult.Err()).NotTo(HaveOccurred())
			Expect(pingResult.Val()).To(Equal("PONG"))
		})

		It("should handle PING with message", func() {
			message := "hello world"
			pingResult := redisClient.Do(ctx, "PING", message)
			Expect(pingResult.Err()).NotTo(HaveOccurred())
			Expect(pingResult.Val()).To(Equal(message))
		})

		It("should handle FLUSHALL command", func() {
			keys := []string{"key1", "key2", "key3"}
			value := "test value"

			for _, key := range keys {
				redisClient.Set(ctx, key, value, 0)
			}

			flushallResult := redisClient.FlushAll(ctx)
			Expect(flushallResult.Err()).NotTo(HaveOccurred())
			Expect(flushallResult.Val()).To(Equal("OK"))

			for _, key := range keys {
				getResult := redisClient.Get(ctx, key)
				Expect(getResult.Err()).To(Equal(redis.Nil))
			}
		})

		It("should handle SEL command for database selection", func() {
			selectResult := redisClient.Do(ctx, "SEL", "0")
			Expect(selectResult.Err()).NotTo(HaveOccurred())
			Expect(selectResult.Val()).To(Equal("OK"))
		})
	})

	Describe("Complex Scenarios", func() {
		It("should handle mixed operations on different data types", func() {
			stringKey := "test:mixed:string"
			listKey := "test:mixed:list"
			setKey := "test:mixed:set"
			zsetKey := "test:mixed:zset"

			redisClient.Set(ctx, stringKey, "string_value", 0)
			redisClient.LPush(ctx, listKey, "list_value")
			redisClient.SAdd(ctx, setKey, "set_value")
			redisClient.ZAdd(ctx, zsetKey, redis.Z{Score: 1.0, Member: "zset_value"})

			stringResult := redisClient.Get(ctx, stringKey)
			Expect(stringResult.Val()).To(Equal("string_value"))

			listResult := redisClient.LRange(ctx, listKey, 0, -1)
			Expect(listResult.Val()).To(Equal([]string{"list_value"}))

			setResult := redisClient.SMembers(ctx, setKey)
			Expect(setResult.Val()).To(Equal([]string{"set_value"}))

			zsetResult := redisClient.ZRange(ctx, zsetKey, 0, -1)
			Expect(zsetResult.Val()).To(Equal([]string{"zset_value"}))
		})

		It("should handle operations with expiration", func() {
			key := "test:expiration:key"
			value := "expiring_value"

			setResult := redisClient.Do(ctx, "SET", key, value, "EX", "1")
			Expect(setResult.Err()).NotTo(HaveOccurred())
			Expect(setResult.Val()).To(Equal("OK"))

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(value))

			time.Sleep(1100 * time.Millisecond)

			getResult = redisClient.Get(ctx, key)
			Expect(getResult.Err()).To(Equal(redis.Nil))
		})

		It("should handle large batch operations", func() {
			const batchSize = 1000
			keyPrefix := "test:batch"

			for i := range batchSize {
				key := fmt.Sprintf("%s:%d", keyPrefix, i)
				value := fmt.Sprintf("value_%d", i)

				result := redisClient.Set(ctx, key, value, 0)
				Expect(result.Err()).NotTo(HaveOccurred())
			}

			for i := range batchSize {
				key := fmt.Sprintf("%s:%d", keyPrefix, i)
				expectedValue := fmt.Sprintf("value_%d", i)

				result := redisClient.Get(ctx, key)
				Expect(result.Err()).NotTo(HaveOccurred())
				Expect(result.Val()).To(Equal(expectedValue))
			}
		})
	})

	Describe("Edge Cases and Error Handling", func() {
		It("should handle operations on non-existent keys", func() {
			nonExistentKey := "test:non:existent"

			getResult := redisClient.Get(ctx, nonExistentKey)
			Expect(getResult.Err()).To(Equal(redis.Nil))

			delResult := redisClient.Del(ctx, nonExistentKey)
			Expect(delResult.Err()).NotTo(HaveOccurred())
			Expect(delResult.Val()).To(Equal(int64(0)))

			existsResult := redisClient.Exists(ctx, nonExistentKey)
			Expect(existsResult.Err()).NotTo(HaveOccurred())
			Expect(existsResult.Val()).To(Equal(int64(0)))
		})

		It("should handle type mismatches gracefully", func() {
			stringKey := "test:type:string"
			redisClient.Set(ctx, stringKey, "string_value", 0)

			lpushResult := redisClient.LPush(ctx, stringKey, "list_value")
			Expect(lpushResult.Err()).To(HaveOccurred())
		})

		It("should handle numeric operations on non-numeric values", func() {
			key := "test:non:numeric"
			redisClient.Set(ctx, key, "not_a_number", 0)

			incrResult := redisClient.Incr(ctx, key)
			Expect(incrResult.Err()).To(HaveOccurred())
		})

		It("should handle empty values", func() {
			key := "test:empty:value"

			setResult := redisClient.Do(ctx, "SET", key, "")
			Expect(setResult.Err()).NotTo(HaveOccurred())
			Expect(setResult.Val()).To(Equal("OK"))

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(""))
		})
	})

	Describe("Performance and Stress Tests", func() {
		It("should handle concurrent operations from multiple clients", func() {
			const numClients = 10
			const operationsPerClient = 100

			clients := make([]*redis.Client, numClients)
			for i := range numClients {
				clients[i] = createRedisClient("localhost:" + testPort)
			}

			done := make(chan bool, numClients)

			for clientID := range numClients {
				go func(id int) {
					defer GinkgoRecover()
					client := clients[id]

					for i := range operationsPerClient {
						key := fmt.Sprintf("concurrent:%d:%d", id, i)
						value := fmt.Sprintf("value_%d_%d", id, i)

						setResult := client.Set(ctx, key, value, 0)
						Expect(setResult.Err()).NotTo(HaveOccurred())

						getResult := client.Get(ctx, key)
						Expect(getResult.Err()).NotTo(HaveOccurred())
						Expect(getResult.Val()).To(Equal(value))
					}

					client.Close()
					done <- true
				}(clientID)
			}

			for range numClients {
				Eventually(done, "30s").Should(Receive())
			}
		})

		It("should handle rapid sequential operations", func() {
			key := "test:rapid:operations"
			const numOperations = 10000

			start := time.Now()

			for i := range numOperations {
				value := strconv.Itoa(i)
				result := redisClient.Set(ctx, key, value, 0)
				Expect(result.Err()).NotTo(HaveOccurred())
			}

			duration := time.Since(start)
			fmt.Printf("Completed %d SET operations in %v (%.2f ops/sec)\n", numOperations, duration, float64(numOperations)/duration.Seconds())

			getResult := redisClient.Get(ctx, key)
			Expect(getResult.Err()).NotTo(HaveOccurred())
			Expect(getResult.Val()).To(Equal(strconv.Itoa(numOperations - 1)))
		})
	})
})
