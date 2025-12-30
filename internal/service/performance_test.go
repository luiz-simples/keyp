package service_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/luiz-simples/keyp.git/internal/service"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Handler Performance Tests", func() {
	var (
		handler     *service.Handler
		storageImpl *storage.Client
		ctx         context.Context
		testDir     string
	)

	BeforeEach(func() {
		ctx = context.Background()

		testDir = createUniqueTestDir("performance")

		var err error
		storageImpl, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		handler = service.NewHandler(storageImpl)
	})

	AfterEach(func() {
		if storageImpl != nil {
			storageImpl.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("Performance Tests", func() {
		It("should handle SET operations efficiently", func() {
			key := []byte("performance:set:key")
			value := []byte("performance set value")
			args := [][]byte{[]byte("SET"), key, value}

			start := time.Now()
			for i := 0; i < 1000; i++ {
				results := handler.Apply(ctx, args)
				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
			}
			duration := time.Since(start)

			Expect(duration).To(BeNumerically("<", time.Second), "SET operations should complete within 1 second")
		})

		It("should handle GET operations efficiently", func() {
			key := []byte("performance:get:key")
			value := []byte("performance get value")

			setArgs := [][]byte{[]byte("SET"), key, value}
			handler.Apply(ctx, setArgs)

			getArgs := [][]byte{[]byte("GET"), key}

			start := time.Now()
			for i := 0; i < 1000; i++ {
				results := handler.Apply(ctx, getArgs)
				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(value))
			}
			duration := time.Since(start)

			Expect(duration).To(BeNumerically("<", time.Second), "GET operations should complete within 1 second")
		})

		It("should handle PING operations efficiently", func() {
			args := [][]byte{[]byte("PING")}

			start := time.Now()
			for i := 0; i < 10000; i++ {
				results := handler.Apply(ctx, args)
				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal([]byte("PONG")))
			}
			duration := time.Since(start)

			Expect(duration).To(BeNumerically("<", 500*time.Millisecond), "PING operations should complete within 0.5 seconds")
		})

		It("should handle mixed operations efficiently", func() {
			baseKey := "performance:mixed"
			value := []byte("mixed operation value")

			start := time.Now()
			for i := 0; i < 1000; i++ {
				key := []byte(fmt.Sprintf("%s:%d", baseKey, i))

				setArgs := [][]byte{[]byte("SET"), key, value}
				setResults := handler.Apply(ctx, setArgs)
				Expect(setResults).To(HaveLen(1))
				Expect(setResults[0].Error).To(BeNil())

				getArgs := [][]byte{[]byte("GET"), key}
				getResults := handler.Apply(ctx, getArgs)
				Expect(getResults).To(HaveLen(1))
				Expect(getResults[0].Error).To(BeNil())
				Expect(getResults[0].Response).To(Equal(value))

				delArgs := [][]byte{[]byte("DEL"), key}
				delResults := handler.Apply(ctx, delArgs)
				Expect(delResults).To(HaveLen(1))
				Expect(delResults[0].Error).To(BeNil())
			}
			duration := time.Since(start)

			Expect(duration).To(BeNumerically("<", 3*time.Second), "Mixed operations should complete within 3 seconds")
		})

		It("should handle large datasets efficiently", func() {
			const numKeys = 1000
			baseKey := "performance:large"
			value := []byte("large dataset test value")

			start := time.Now()
			for i := 0; i < numKeys; i++ {
				key := []byte(fmt.Sprintf("%s:%d", baseKey, i))
				setArgs := [][]byte{[]byte("SET"), key, value}
				results := handler.Apply(ctx, setArgs)
				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
			}
			setupDuration := time.Since(start)

			start = time.Now()
			for i := 0; i < numKeys; i++ {
				key := []byte(fmt.Sprintf("%s:%d", baseKey, i))
				getArgs := [][]byte{[]byte("GET"), key}
				results := handler.Apply(ctx, getArgs)
				Expect(results).To(HaveLen(1))
				Expect(results[0].Error).To(BeNil())
				Expect(results[0].Response).To(Equal(value))
			}
			readDuration := time.Since(start)

			Expect(setupDuration).To(BeNumerically("<", 2*time.Second), "Setup should complete within 2 seconds")
			Expect(readDuration).To(BeNumerically("<", 2*time.Second), "Read should complete within 2 seconds")
		})
	})
})

func BenchmarkHandlerSET(b *testing.B) {
	ctx := context.Background()
	testDir := createUniqueTestDir("bench-set")
	defer cleanupTestDir(testDir)

	storageImpl, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storageImpl.Close()

	handler := service.NewHandler(storageImpl)
	key := []byte("benchmark:key")
	value := []byte("benchmark value")
	args := [][]byte{[]byte("SET"), key, value}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := handler.Apply(ctx, args)
		if len(results) != 1 || results[0].Error != nil {
			b.Fatal("SET operation failed")
		}
	}
}

func BenchmarkHandlerGET(b *testing.B) {
	ctx := context.Background()
	testDir := createUniqueTestDir("bench-get")
	defer cleanupTestDir(testDir)

	storageImpl, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storageImpl.Close()

	handler := service.NewHandler(storageImpl)
	key := []byte("benchmark:key")
	value := []byte("benchmark value")

	setArgs := [][]byte{[]byte("SET"), key, value}
	handler.Apply(ctx, setArgs)

	getArgs := [][]byte{[]byte("GET"), key}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := handler.Apply(ctx, getArgs)
		if len(results) != 1 || results[0].Error != nil {
			b.Fatal("GET operation failed")
		}
	}
}

func BenchmarkHandlerDEL(b *testing.B) {
	ctx := context.Background()
	testDir := createUniqueTestDir("bench-del")
	defer cleanupTestDir(testDir)

	storageImpl, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storageImpl.Close()

	handler := service.NewHandler(storageImpl)
	value := []byte("benchmark value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("benchmark:key:%d", i))

		setArgs := [][]byte{[]byte("SET"), key, value}
		handler.Apply(ctx, setArgs)

		delArgs := [][]byte{[]byte("DEL"), key}
		results := handler.Apply(ctx, delArgs)
		if len(results) != 1 || results[0].Error != nil {
			b.Fatal("DEL operation failed")
		}
	}
}

func BenchmarkHandlerPING(b *testing.B) {
	ctx := context.Background()
	testDir := createUniqueTestDir("bench-ping")
	defer cleanupTestDir(testDir)

	storageImpl, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storageImpl.Close()

	handler := service.NewHandler(storageImpl)
	args := [][]byte{[]byte("PING")}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		results := handler.Apply(ctx, args)
		if len(results) != 1 || results[0].Error != nil {
			b.Fatal("PING operation failed")
		}
	}
}

func BenchmarkHandlerMixed(b *testing.B) {
	ctx := context.Background()
	testDir := createUniqueTestDir("bench-mixed")
	defer cleanupTestDir(testDir)

	storageImpl, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer storageImpl.Close()

	handler := service.NewHandler(storageImpl)
	value := []byte("benchmark value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := []byte(fmt.Sprintf("benchmark:mixed:%d", i))

		setArgs := [][]byte{[]byte("SET"), key, value}
		setResults := handler.Apply(ctx, setArgs)
		if len(setResults) != 1 || setResults[0].Error != nil {
			b.Fatal("SET operation failed")
		}

		getArgs := [][]byte{[]byte("GET"), key}
		getResults := handler.Apply(ctx, getArgs)
		if len(getResults) != 1 || getResults[0].Error != nil {
			b.Fatal("GET operation failed")
		}

		delArgs := [][]byte{[]byte("DEL"), key}
		delResults := handler.Apply(ctx, delArgs)
		if len(delResults) != 1 || delResults[0].Error != nil {
			b.Fatal("DEL operation failed")
		}
	}
}
