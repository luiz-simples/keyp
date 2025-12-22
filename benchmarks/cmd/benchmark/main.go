package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type BenchmarkConfig struct {
	ServerType    string `json:"server_type"`
	Address       string `json:"address"`
	NumOperations int    `json:"num_operations"`
	NumClients    int    `json:"num_clients"`
	KeySize       int    `json:"key_size"`
	ValueSize     int    `json:"value_size"`
	TTLSeconds    int    `json:"ttl_seconds"`
}

type CommandResult struct {
	Command      string        `json:"command"`
	TotalOps     int           `json:"total_ops"`
	Duration     time.Duration `json:"duration"`
	OpsPerSecond float64       `json:"ops_per_second"`
	AvgLatency   time.Duration `json:"avg_latency"`
	P95Latency   time.Duration `json:"p95_latency"`
	P99Latency   time.Duration `json:"p99_latency"`
	MinLatency   time.Duration `json:"min_latency"`
	MaxLatency   time.Duration `json:"max_latency"`
	ErrorCount   int           `json:"error_count"`
	SuccessRate  float64       `json:"success_rate"`
}

type SystemMetrics struct {
	MemoryUsageMB float64   `json:"memory_usage_mb"`
	CPUUsage      float64   `json:"cpu_usage"`
	Timestamp     time.Time `json:"timestamp"`
}

type BenchmarkResult struct {
	Config        BenchmarkConfig `json:"config"`
	Commands      []CommandResult `json:"commands"`
	SystemMetrics SystemMetrics   `json:"system_metrics"`
	StartTime     time.Time       `json:"start_time"`
	EndTime       time.Time       `json:"end_time"`
	TotalDuration time.Duration   `json:"total_duration"`
}

func main() {
	var (
		serverType = flag.String("server", "keyp", "Server type: keyp or redis")
		address    = flag.String("addr", "localhost:6380", "Server address")
		numOps     = flag.Int("ops", 10000, "Number of operations per command")
		numClients = flag.Int("clients", 10, "Number of concurrent clients")
		keySize    = flag.Int("keysize", 16, "Key size in bytes")
		valueSize  = flag.Int("valuesize", 64, "Value size in bytes")
		ttlSeconds = flag.Int("ttl", 300, "TTL in seconds for EXPIRE tests")
		outputDir  = flag.String("output", "", "Output directory for results")
	)
	flag.Parse()

	if *outputDir == "" {
		*outputDir = filepath.Join("benchmarks", "results", time.Now().Format("2006-01-02"), *serverType)
	}

	config := BenchmarkConfig{
		ServerType:    *serverType,
		Address:       *address,
		NumOperations: *numOps,
		NumClients:    *numClients,
		KeySize:       *keySize,
		ValueSize:     *valueSize,
		TTLSeconds:    *ttlSeconds,
	}

	fmt.Printf("Starting benchmark for %s server at %s\n", config.ServerType, config.Address)
	fmt.Printf("Operations: %d, Clients: %d, Key size: %d, Value size: %d\n",
		config.NumOperations, config.NumClients, config.KeySize, config.ValueSize)

	client := redis.NewClient(&redis.Options{
		Addr: config.Address,
	})
	defer client.Close()

	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		log.Fatalf("Failed to connect to server: %v", err)
	}

	result := BenchmarkResult{
		Config:    config,
		StartTime: time.Now(),
	}

	commands := []string{"SET", "GET", "DEL", "EXPIRE", "TTL", "PERSIST"}

	for _, cmd := range commands {
		fmt.Printf("\nRunning %s benchmark...\n", cmd)
		cmdResult := runCommandBenchmark(ctx, client, cmd, config)
		result.Commands = append(result.Commands, cmdResult)

		fmt.Printf("%s: %.2f ops/sec, avg latency: %v\n",
			cmd, cmdResult.OpsPerSecond, cmdResult.AvgLatency)
	}

	result.EndTime = time.Now()
	result.TotalDuration = result.EndTime.Sub(result.StartTime)
	result.SystemMetrics = getSystemMetrics()

	if err := saveResults(result, *outputDir); err != nil {
		log.Fatalf("Failed to save results: %v", err)
	}

	fmt.Printf("\nBenchmark completed in %v\n", result.TotalDuration)
	fmt.Printf("Results saved to: %s\n", *outputDir)
}

func runCommandBenchmark(ctx context.Context, client *redis.Client, command string, config BenchmarkConfig) CommandResult {
	var wg sync.WaitGroup
	var mu sync.Mutex

	latencies := make([]time.Duration, 0, config.NumOperations)
	errorCount := 0

	opsPerClient := config.NumOperations / config.NumClients

	startTime := time.Now()

	for i := 0; i < config.NumClients; i++ {
		wg.Add(1)
		go func(clientID int) {
			defer wg.Done()

			for j := 0; j < opsPerClient; j++ {
				key := fmt.Sprintf("bench:%s:%d:%d", command, clientID, j)
				value := generateValue(config.ValueSize)

				opStart := time.Now()
				var err error

				switch command {
				case "SET":
					err = client.Set(ctx, key, value, 0).Err()
				case "GET":
					client.Set(ctx, key, value, 0)
					_, err = client.Get(ctx, key).Result()
				case "DEL":
					client.Set(ctx, key, value, 0)
					err = client.Del(ctx, key).Err()
				case "EXPIRE":
					client.Set(ctx, key, value, 0)
					err = client.Expire(ctx, key, time.Duration(config.TTLSeconds)*time.Second).Err()
				case "TTL":
					client.Set(ctx, key, value, 0)
					client.Expire(ctx, key, time.Duration(config.TTLSeconds)*time.Second)
					_, err = client.TTL(ctx, key).Result()
				case "PERSIST":
					client.Set(ctx, key, value, 0)
					client.Expire(ctx, key, time.Duration(config.TTLSeconds)*time.Second)
					err = client.Persist(ctx, key).Err()
				}

				latency := time.Since(opStart)

				mu.Lock()
				latencies = append(latencies, latency)
				if err != nil {
					errorCount++
				}
				mu.Unlock()
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	return calculateResults(command, latencies, duration, errorCount, config.NumOperations)
}

func calculateResults(command string, latencies []time.Duration, duration time.Duration, errorCount, totalOps int) CommandResult {
	if len(latencies) == 0 {
		return CommandResult{
			Command:     command,
			TotalOps:    totalOps,
			Duration:    duration,
			ErrorCount:  errorCount,
			SuccessRate: 0,
		}
	}

	var totalLatency time.Duration
	minLatency := latencies[0]
	maxLatency := latencies[0]

	for _, lat := range latencies {
		totalLatency += lat
		if lat < minLatency {
			minLatency = lat
		}
		if lat > maxLatency {
			maxLatency = lat
		}
	}

	avgLatency := totalLatency / time.Duration(len(latencies))

	sortedLatencies := make([]time.Duration, len(latencies))
	copy(sortedLatencies, latencies)

	for i := 0; i < len(sortedLatencies)-1; i++ {
		for j := i + 1; j < len(sortedLatencies); j++ {
			if sortedLatencies[i] > sortedLatencies[j] {
				sortedLatencies[i], sortedLatencies[j] = sortedLatencies[j], sortedLatencies[i]
			}
		}
	}

	p95Index := int(float64(len(sortedLatencies)) * 0.95)
	p99Index := int(float64(len(sortedLatencies)) * 0.99)

	if p95Index >= len(sortedLatencies) {
		p95Index = len(sortedLatencies) - 1
	}
	if p99Index >= len(sortedLatencies) {
		p99Index = len(sortedLatencies) - 1
	}

	successOps := totalOps - errorCount
	opsPerSecond := float64(successOps) / duration.Seconds()
	successRate := float64(successOps) / float64(totalOps) * 100

	return CommandResult{
		Command:      command,
		TotalOps:     totalOps,
		Duration:     duration,
		OpsPerSecond: opsPerSecond,
		AvgLatency:   avgLatency,
		P95Latency:   sortedLatencies[p95Index],
		P99Latency:   sortedLatencies[p99Index],
		MinLatency:   minLatency,
		MaxLatency:   maxLatency,
		ErrorCount:   errorCount,
		SuccessRate:  successRate,
	}
}

func generateValue(size int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, size)
	for i := range b {
		b[i] = charset[i%len(charset)]
	}
	return string(b)
}

func getSystemMetrics() SystemMetrics {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return SystemMetrics{
		MemoryUsageMB: float64(m.Alloc) / 1024 / 1024,
		CPUUsage:      0.0,
		Timestamp:     time.Now(),
	}
}

func saveResults(result BenchmarkResult, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	timestamp := result.StartTime.Format("15-04-05")
	filename := filepath.Join(outputDir, fmt.Sprintf("benchmark_%s.json", timestamp))

	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}
