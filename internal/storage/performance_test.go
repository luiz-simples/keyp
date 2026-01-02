package storage_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gmeasure"

	"github.com/luiz-simples/keyp.git/internal/domain"
	"github.com/luiz-simples/keyp.git/internal/storage"
)

var _ = Describe("Storage Performance Tests", Label("performance"), func() {
	var (
		client  *storage.Client
		testDir string
		ctx     context.Context
	)

	BeforeEach(func() {
		testDir = createUniqueTestDir("performance")
		var err error
		client, err = storage.NewClient(testDir)
		Expect(err).NotTo(HaveOccurred())
		Expect(client).NotTo(BeNil())

		ctx = context.WithValue(context.Background(), domain.DB, uint8(0))
	})

	AfterEach(func() {
		if client != nil {
			client.Close()
		}
		cleanupTestDir(testDir)
	})

	Describe("Operation Performance", func() {
		Context("when measuring individual operations", func() {
			It("should perform SET operations efficiently", func() {
				experiment := gmeasure.NewExperiment("SET Performance")
				AddReportEntry(experiment.Name, experiment)

				key := []byte("benchmark-set-key")
				value := []byte("benchmark-set-value")

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("SET", func() {
						err := client.Set(ctx, key, value)
						Expect(err).NotTo(HaveOccurred())
					})
				}, gmeasure.SamplingConfig{N: 1000})

				Expect(experiment.GetStats("SET").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", time.Millisecond))
			})

			It("should perform GET operations efficiently", func() {
				experiment := gmeasure.NewExperiment("GET Performance")
				AddReportEntry(experiment.Name, experiment)

				key := []byte("benchmark-get-key")
				value := []byte("benchmark-get-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("GET", func() {
						result, err := client.Get(ctx, key)
						Expect(err).NotTo(HaveOccurred())
						Expect(result).To(Equal(value))
					})
				}, gmeasure.SamplingConfig{N: 1000})

				Expect(experiment.GetStats("GET").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", time.Millisecond))
			})

			It("should perform DEL operations efficiently", func() {
				experiment := gmeasure.NewExperiment("DEL Performance")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					key := []byte(fmt.Sprintf("benchmark-del-key-%d", idx))
					value := []byte("benchmark-del-value")

					err := client.Set(ctx, key, value)
					Expect(err).NotTo(HaveOccurred())

					experiment.MeasureDuration("DEL", func() {
						deleted, err := client.Del(ctx, key)
						Expect(err).NotTo(HaveOccurred())
						Expect(deleted).To(Equal(uint32(1)))
					})
				}, gmeasure.SamplingConfig{N: 1000})

				Expect(experiment.GetStats("DEL").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 5*time.Millisecond))
			})

			It("should perform TTL operations efficiently", func() {
				experiment := gmeasure.NewExperiment("TTL Performance")
				AddReportEntry(experiment.Name, experiment)

				key := []byte("benchmark-ttl-key")
				value := []byte("benchmark-ttl-value")

				err := client.Set(ctx, key, value)
				Expect(err).NotTo(HaveOccurred())

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("TTL", func() {
						client.Expire(ctx, key, 3600)
						ttl := client.TTL(ctx, key)
						Expect(ttl).To(BeNumerically(">", 0))
					})
				}, gmeasure.SamplingConfig{N: 1000})

				Expect(experiment.GetStats("TTL").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", time.Millisecond))
			})
		})

		Context("when measuring batch operations", func() {
			It("should handle batch SET operations efficiently", func() {
				experiment := gmeasure.NewExperiment("Batch SET Performance")
				AddReportEntry(experiment.Name, experiment)

				const batchSize = 100

				experiment.Sample(func(idx int) {
					keys := make([][]byte, batchSize)
					values := make([][]byte, batchSize)

					for i := range batchSize {
						keys[i] = []byte(fmt.Sprintf("batch-set-key-%d-%d", idx, i))
						values[i] = []byte(fmt.Sprintf("batch-set-value-%d-%d", idx, i))
					}

					experiment.MeasureDuration("BatchSET", func() {
						for i := range batchSize {
							err := client.Set(ctx, keys[i], values[i])
							Expect(err).NotTo(HaveOccurred())
						}
					})
				}, gmeasure.SamplingConfig{N: 100})

				avgTime := experiment.GetStats("BatchSET").DurationFor(gmeasure.StatMean) / batchSize
				Expect(avgTime).To(BeNumerically("<", time.Millisecond))
			})

			It("should handle batch GET operations efficiently", func() {
				experiment := gmeasure.NewExperiment("Batch GET Performance")
				AddReportEntry(experiment.Name, experiment)

				const batchSize = 100
				keys := make([][]byte, batchSize)
				values := make([][]byte, batchSize)

				for i := range batchSize {
					keys[i] = []byte(fmt.Sprintf("batch-get-key-%d", i))
					values[i] = []byte(fmt.Sprintf("batch-get-value-%d", i))
					err := client.Set(ctx, keys[i], values[i])
					Expect(err).NotTo(HaveOccurred())
				}

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("BatchGET", func() {
						for i := range batchSize {
							result, err := client.Get(ctx, keys[i])
							Expect(err).NotTo(HaveOccurred())
							Expect(result).To(Equal(values[i]))
						}
					})
				}, gmeasure.SamplingConfig{N: 100})

				avgTime := experiment.GetStats("BatchGET").DurationFor(gmeasure.StatMean) / batchSize
				Expect(avgTime).To(BeNumerically("<", time.Millisecond))
			})

			It("should handle batch DEL operations efficiently", func() {
				experiment := gmeasure.NewExperiment("Batch DEL Performance")
				AddReportEntry(experiment.Name, experiment)

				experiment.Sample(func(idx int) {
					const batchSize = 100
					keys := make([][]byte, batchSize)

					for i := range batchSize {
						keys[i] = []byte(fmt.Sprintf("batch-del-key-%d-%d", idx, i))
						err := client.Set(ctx, keys[i], []byte("value"))
						Expect(err).NotTo(HaveOccurred())
					}

					experiment.MeasureDuration("BatchDEL", func() {
						deleted, err := client.Del(ctx, keys...)
						Expect(err).NotTo(HaveOccurred())
						Expect(deleted).To(Equal(uint32(batchSize)))
					})
				}, gmeasure.SamplingConfig{N: 100})

				Expect(experiment.GetStats("BatchDEL").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 100*time.Millisecond))
			})
		})

		Context("when measuring large data operations", func() {
			It("should handle large values efficiently", func() {
				experiment := gmeasure.NewExperiment("Large Value Performance")
				AddReportEntry(experiment.Name, experiment)

				key := []byte("large-value-key")
				largeValue := make([]byte, 1024*100)
				for i := range largeValue {
					largeValue[i] = byte(i % 256)
				}

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("LargeSET", func() {
						err := client.Set(ctx, key, largeValue)
						Expect(err).NotTo(HaveOccurred())
					})

					experiment.MeasureDuration("LargeGET", func() {
						result, err := client.Get(ctx, key)
						Expect(err).NotTo(HaveOccurred())
						Expect(len(result)).To(Equal(len(largeValue)))
					})
				}, gmeasure.SamplingConfig{N: 100})

				Expect(experiment.GetStats("LargeSET").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Millisecond))
				Expect(experiment.GetStats("LargeGET").DurationFor(gmeasure.StatMean)).To(BeNumerically("<", 10*time.Millisecond))
			})
		})

		Context("when measuring concurrent operations", func() {
			It("should maintain performance under concurrent load", func() {
				experiment := gmeasure.NewExperiment("Concurrent Performance")
				AddReportEntry(experiment.Name, experiment)

				const numGoroutines = 10
				const opsPerGoroutine = 100

				experiment.Sample(func(idx int) {
					experiment.MeasureDuration("ConcurrentOps", func() {
						done := make(chan bool, numGoroutines)

						for i := range numGoroutines {
							go func(goroutineID int) {
								defer GinkgoRecover()
								for j := range opsPerGoroutine {
									key := []byte(fmt.Sprintf("concurrent-key-%d-%d-%d", idx, goroutineID, j))
									value := []byte(fmt.Sprintf("concurrent-value-%d-%d-%d", idx, goroutineID, j))

									err := client.Set(ctx, key, value)
									Expect(err).NotTo(HaveOccurred())

									result, err := client.Get(ctx, key)
									Expect(err).NotTo(HaveOccurred())
									Expect(result).To(Equal(value))
								}
								done <- true
							}(i)
						}

						for range numGoroutines {
							<-done
						}
					})
				}, gmeasure.SamplingConfig{N: 50})

				totalOps := numGoroutines * opsPerGoroutine * 2
				avgTimePerOp := experiment.GetStats("ConcurrentOps").DurationFor(gmeasure.StatMean) / time.Duration(totalOps)
				Expect(avgTimePerOp).To(BeNumerically("<", time.Millisecond))
			})
		})
	})

	Describe("Memory Usage Tests", func() {
		Context("when measuring memory efficiency", func() {
			It("should handle many small keys efficiently", func() {
				const numKeys = 10000
				keyPrefix := "memory-test-key-"
				value := []byte("small-value")

				start := time.Now()

				for i := range numKeys {
					key := []byte(fmt.Sprintf("%s%d", keyPrefix, i))
					err := client.Set(ctx, key, value)
					Expect(err).NotTo(HaveOccurred())
				}

				setDuration := time.Since(start)
				avgSetTime := setDuration.Nanoseconds() / int64(numKeys)

				start = time.Now()

				for i := range numKeys {
					key := []byte(fmt.Sprintf("%s%d", keyPrefix, i))
					result, err := client.Get(ctx, key)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(value))
				}

				getDuration := time.Since(start)
				avgGetTime := getDuration.Nanoseconds() / int64(numKeys)

				Expect(avgSetTime).To(BeNumerically("<", 50000), "Average SET time should be less than 50μs")
				Expect(avgGetTime).To(BeNumerically("<", 50000), "Average GET time should be less than 50μs")
			})

			It("should handle large values efficiently", func() {
				const numLargeValues = 100
				const valueSize = 1024 * 10

				largeValues := make([][]byte, numLargeValues)
				for i := range numLargeValues {
					largeValues[i] = make([]byte, valueSize)
					for j := range largeValues[i] {
						largeValues[i][j] = byte((i + j) % 256)
					}
				}

				start := time.Now()

				for i := range numLargeValues {
					key := []byte(fmt.Sprintf("large-key-%d", i))
					err := client.Set(ctx, key, largeValues[i])
					Expect(err).NotTo(HaveOccurred())
				}

				setDuration := time.Since(start)
				avgSetTime := setDuration.Nanoseconds() / int64(numLargeValues)

				start = time.Now()

				for i := range numLargeValues {
					key := []byte(fmt.Sprintf("large-key-%d", i))
					result, err := client.Get(ctx, key)
					Expect(err).NotTo(HaveOccurred())
					Expect(len(result)).To(Equal(valueSize))
				}

				getDuration := time.Since(start)
				avgGetTime := getDuration.Nanoseconds() / int64(numLargeValues)

				Expect(avgSetTime).To(BeNumerically("<", 500000), "Average large SET time should be less than 500μs")
				Expect(avgGetTime).To(BeNumerically("<", 500000), "Average large GET time should be less than 500μs")
			})
		})
	})

	Describe("Throughput Tests", func() {
		Context("when measuring sustained throughput", func() {
			It("should maintain performance under sustained load", func() {
				const duration = 5 * time.Second
				const targetOpsPerSecond = 10000

				operationCount := 0
				start := time.Now()

				for time.Since(start) < duration {
					key := []byte(fmt.Sprintf("throughput-key-%d", operationCount))
					value := []byte(fmt.Sprintf("throughput-value-%d", operationCount))

					err := client.Set(ctx, key, value)
					Expect(err).NotTo(HaveOccurred())

					result, err := client.Get(ctx, key)
					Expect(err).NotTo(HaveOccurred())
					Expect(result).To(Equal(value))

					operationCount += 2
				}

				actualDuration := time.Since(start)
				actualOpsPerSecond := float64(operationCount) / actualDuration.Seconds()

				Expect(actualOpsPerSecond).To(BeNumerically(">", float64(targetOpsPerSecond)*0.5),
					fmt.Sprintf("Should achieve at least 50%% of target throughput (%d ops/sec)", targetOpsPerSecond))
			})
		})
	})
})

func BenchmarkStorageSet(b *testing.B) {
	testDir := createUniqueTestDir("bench-set")
	defer cleanupTestDir(testDir)

	client, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.WithValue(context.Background(), domain.DB, uint8(0))
	key := []byte("benchmark-key")
	value := []byte("benchmark-value")

	b.ResetTimer()
	for range b.N {
		err := client.Set(ctx, key, value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStorageGet(b *testing.B) {
	testDir := createUniqueTestDir("bench-get")
	defer cleanupTestDir(testDir)

	client, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.WithValue(context.Background(), domain.DB, uint8(0))
	key := []byte("benchmark-key")
	value := []byte("benchmark-value")

	err = client.Set(ctx, key, value)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for range b.N {
		_, err := client.Get(ctx, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStorageDel(b *testing.B) {
	testDir := createUniqueTestDir("bench-del")
	defer cleanupTestDir(testDir)

	client, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.WithValue(context.Background(), domain.DB, uint8(0))

	b.ResetTimer()
	for i := range b.N {
		key := []byte(fmt.Sprintf("benchmark-key-%d", i))
		value := []byte("benchmark-value")

		err := client.Set(ctx, key, value)
		if err != nil {
			b.Fatal(err)
		}

		_, err = client.Del(ctx, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkStorageMixed(b *testing.B) {
	testDir := createUniqueTestDir("bench-mixed")
	defer cleanupTestDir(testDir)

	client, err := storage.NewClient(testDir)
	if err != nil {
		b.Fatal(err)
	}
	defer client.Close()

	ctx := context.WithValue(context.Background(), domain.DB, uint8(0))

	b.ResetTimer()
	for i := range b.N {
		key := []byte(fmt.Sprintf("mixed-key-%d", i))
		value := []byte(fmt.Sprintf("mixed-value-%d", i))

		err := client.Set(ctx, key, value)
		if err != nil {
			b.Fatal(err)
		}

		_, err = client.Get(ctx, key)
		if err != nil {
			b.Fatal(err)
		}

		_, err = client.Del(ctx, key)
		if err != nil {
			b.Fatal(err)
		}
	}
}
