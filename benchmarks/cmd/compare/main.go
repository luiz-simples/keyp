package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
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

type ComparisonRow struct {
	Command          string  `json:"command"`
	RedisOpsPerSec   float64 `json:"redis_ops_per_sec"`
	KeypOpsPerSec    float64 `json:"keyp_ops_per_sec"`
	PerformanceRatio float64 `json:"performance_ratio"`
	RedisAvgLatency  string  `json:"redis_avg_latency"`
	KeypAvgLatency   string  `json:"keyp_avg_latency"`
	RedisP95Latency  string  `json:"redis_p95_latency"`
	KeypP95Latency   string  `json:"keyp_p95_latency"`
	RedisP99Latency  string  `json:"redis_p99_latency"`
	KeypP99Latency   string  `json:"keyp_p99_latency"`
	RedisSuccessRate float64 `json:"redis_success_rate"`
	KeypSuccessRate  float64 `json:"keyp_success_rate"`
}

type ComparisonReport struct {
	GeneratedAt time.Time       `json:"generated_at"`
	RedisConfig BenchmarkConfig `json:"redis_config"`
	KeypConfig  BenchmarkConfig `json:"keyp_config"`
	Commands    []ComparisonRow `json:"commands"`
	Summary     Summary         `json:"summary"`
}

type Summary struct {
	OverallKeypPerformance string  `json:"overall_keyp_performance"`
	BestKeypCommand        string  `json:"best_keyp_command"`
	WorstKeypCommand       string  `json:"worst_keyp_command"`
	AvgPerformanceRatio    float64 `json:"avg_performance_ratio"`
}

func main() {
	var (
		redisDir  = flag.String("redis", "", "Redis results directory")
		keypDir   = flag.String("keyp", "", "Keyp results directory")
		outputDir = flag.String("output", "", "Output directory for comparison")
	)
	flag.Parse()

	if *redisDir == "" || *keypDir == "" || *outputDir == "" {
		log.Fatal("All directories (redis, keyp, output) must be specified")
	}

	fmt.Println("📊 Generating comparison report...")

	redisResult, err := loadLatestResult(*redisDir)
	if err != nil {
		log.Fatalf("Failed to load Redis results: %v", err)
	}

	keypResult, err := loadLatestResult(*keypDir)
	if err != nil {
		log.Fatalf("Failed to load Keyp results: %v", err)
	}

	report := generateComparison(redisResult, keypResult)

	if err := saveReport(report, *outputDir); err != nil {
		log.Fatalf("Failed to save report: %v", err)
	}

	printSummary(report)
	fmt.Printf("📁 Detailed report saved to: %s\n", *outputDir)
}

func loadLatestResult(dir string) (*BenchmarkResult, error) {
	var latestFile string
	var latestTime time.Time

	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".json") {
			info, err := d.Info()
			if err != nil {
				return err
			}

			if info.ModTime().After(latestTime) {
				latestTime = info.ModTime()
				latestFile = path
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if latestFile == "" {
		return nil, fmt.Errorf("no benchmark files found in %s", dir)
	}

	data, err := os.ReadFile(latestFile)
	if err != nil {
		return nil, err
	}

	var result BenchmarkResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func generateComparison(redisResult, keypResult *BenchmarkResult) ComparisonReport {
	report := ComparisonReport{
		GeneratedAt: time.Now(),
		RedisConfig: redisResult.Config,
		KeypConfig:  keypResult.Config,
	}

	redisCommands := make(map[string]CommandResult)
	for _, cmd := range redisResult.Commands {
		redisCommands[cmd.Command] = cmd
	}

	keypCommands := make(map[string]CommandResult)
	for _, cmd := range keypResult.Commands {
		keypCommands[cmd.Command] = cmd
	}

	var totalRatio float64
	var bestRatio, worstRatio float64
	var bestCmd, worstCmd string

	for _, command := range []string{"SET", "GET", "DEL", "EXPIRE", "TTL", "PERSIST"} {
		redisCmd, redisOk := redisCommands[command]
		keypCmd, keypOk := keypCommands[command]

		if !redisOk || !keypOk {
			continue
		}

		ratio := keypCmd.OpsPerSecond / redisCmd.OpsPerSecond

		if bestCmd == "" || ratio > bestRatio {
			bestRatio = ratio
			bestCmd = command
		}

		if worstCmd == "" || ratio < worstRatio {
			worstRatio = ratio
			worstCmd = command
		}

		totalRatio += ratio

		row := ComparisonRow{
			Command:          command,
			RedisOpsPerSec:   redisCmd.OpsPerSecond,
			KeypOpsPerSec:    keypCmd.OpsPerSecond,
			PerformanceRatio: ratio,
			RedisAvgLatency:  formatDuration(redisCmd.AvgLatency),
			KeypAvgLatency:   formatDuration(keypCmd.AvgLatency),
			RedisP95Latency:  formatDuration(redisCmd.P95Latency),
			KeypP95Latency:   formatDuration(keypCmd.P95Latency),
			RedisP99Latency:  formatDuration(redisCmd.P99Latency),
			KeypP99Latency:   formatDuration(keypCmd.P99Latency),
			RedisSuccessRate: redisCmd.SuccessRate,
			KeypSuccessRate:  keypCmd.SuccessRate,
		}

		report.Commands = append(report.Commands, row)
	}

	avgRatio := totalRatio / float64(len(report.Commands))

	var performance string
	if avgRatio >= 1.1 {
		performance = "Superior"
	} else if avgRatio >= 0.9 {
		performance = "Equivalent"
	} else {
		performance = "Inferior"
	}

	report.Summary = Summary{
		OverallKeypPerformance: performance,
		BestKeypCommand:        bestCmd,
		WorstKeypCommand:       worstCmd,
		AvgPerformanceRatio:    avgRatio,
	}

	return report
}

func formatDuration(d time.Duration) string {
	if d < time.Microsecond {
		return fmt.Sprintf("%.0fns", float64(d.Nanoseconds()))
	}
	if d < time.Millisecond {
		return fmt.Sprintf("%.1fμs", float64(d.Nanoseconds())/1000)
	}
	return fmt.Sprintf("%.2fms", float64(d.Nanoseconds())/1000000)
}

func saveReport(report ComparisonReport, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	timestamp := report.GeneratedAt.Format("15-04-05")

	jsonFile := filepath.Join(outputDir, fmt.Sprintf("comparison_%s.json", timestamp))
	jsonData, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(jsonFile, jsonData, 0644); err != nil {
		return err
	}

	markdownFile := filepath.Join(outputDir, fmt.Sprintf("comparison_%s.md", timestamp))
	markdown := generateMarkdownReport(report)
	if err := os.WriteFile(markdownFile, []byte(markdown), 0644); err != nil {
		return err
	}

	csvFile := filepath.Join(outputDir, fmt.Sprintf("comparison_%s.csv", timestamp))
	csv := generateCSVReport(report)
	if err := os.WriteFile(csvFile, []byte(csv), 0644); err != nil {
		return err
	}

	return nil
}

func generateMarkdownReport(report ComparisonReport) string {
	var sb strings.Builder

	sb.WriteString("# Keyp vs Redis Performance Comparison\n\n")
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n\n", report.GeneratedAt.Format("2006-01-02 15:04:05")))

	sb.WriteString("## Configuration\n\n")
	sb.WriteString("| Parameter | Redis | Keyp |\n")
	sb.WriteString("|-----------|-------|------|\n")
	sb.WriteString(fmt.Sprintf("| Operations | %d | %d |\n", report.RedisConfig.NumOperations, report.KeypConfig.NumOperations))
	sb.WriteString(fmt.Sprintf("| Clients | %d | %d |\n", report.RedisConfig.NumClients, report.KeypConfig.NumClients))
	sb.WriteString(fmt.Sprintf("| Key Size | %d bytes | %d bytes |\n", report.RedisConfig.KeySize, report.KeypConfig.KeySize))
	sb.WriteString(fmt.Sprintf("| Value Size | %d bytes | %d bytes |\n", report.RedisConfig.ValueSize, report.KeypConfig.ValueSize))
	sb.WriteString("\n")

	sb.WriteString("## Performance Summary\n\n")
	sb.WriteString(fmt.Sprintf("- **Overall Keyp Performance:** %s\n", report.Summary.OverallKeypPerformance))
	sb.WriteString(fmt.Sprintf("- **Average Performance Ratio:** %.2fx\n", report.Summary.AvgPerformanceRatio))
	sb.WriteString(fmt.Sprintf("- **Best Keyp Command:** %s\n", report.Summary.BestKeypCommand))
	sb.WriteString(fmt.Sprintf("- **Worst Keyp Command:** %s\n", report.Summary.WorstKeypCommand))
	sb.WriteString("\n")

	sb.WriteString("## Detailed Results\n\n")
	sb.WriteString("| Command | Redis ops/sec | Keyp ops/sec | Ratio | Redis Avg | Keyp Avg | Redis P95 | Keyp P95 |\n")
	sb.WriteString("|---------|---------------|--------------|-------|-----------|----------|-----------|----------|\n")

	for _, cmd := range report.Commands {
		sb.WriteString(fmt.Sprintf("| %s | %.0f | %.0f | %.2fx | %s | %s | %s | %s |\n",
			cmd.Command,
			cmd.RedisOpsPerSec,
			cmd.KeypOpsPerSec,
			cmd.PerformanceRatio,
			cmd.RedisAvgLatency,
			cmd.KeypAvgLatency,
			cmd.RedisP95Latency,
			cmd.KeypP95Latency,
		))
	}

	sb.WriteString("\n## Success Rates\n\n")
	sb.WriteString("| Command | Redis Success Rate | Keyp Success Rate |\n")
	sb.WriteString("|---------|-------------------|------------------|\n")

	for _, cmd := range report.Commands {
		sb.WriteString(fmt.Sprintf("| %s | %.1f%% | %.1f%% |\n",
			cmd.Command,
			cmd.RedisSuccessRate,
			cmd.KeypSuccessRate,
		))
	}

	return sb.String()
}

func generateCSVReport(report ComparisonReport) string {
	var sb strings.Builder

	sb.WriteString("Command,Redis_OpsPerSec,Keyp_OpsPerSec,Performance_Ratio,Redis_AvgLatency,Keyp_AvgLatency,Redis_P95Latency,Keyp_P95Latency,Redis_P99Latency,Keyp_P99Latency,Redis_SuccessRate,Keyp_SuccessRate\n")

	for _, cmd := range report.Commands {
		sb.WriteString(fmt.Sprintf("%s,%.2f,%.2f,%.3f,%s,%s,%s,%s,%s,%s,%.2f,%.2f\n",
			cmd.Command,
			cmd.RedisOpsPerSec,
			cmd.KeypOpsPerSec,
			cmd.PerformanceRatio,
			cmd.RedisAvgLatency,
			cmd.KeypAvgLatency,
			cmd.RedisP95Latency,
			cmd.KeypP95Latency,
			cmd.RedisP99Latency,
			cmd.KeypP99Latency,
			cmd.RedisSuccessRate,
			cmd.KeypSuccessRate,
		))
	}

	return sb.String()
}

func printSummary(report ComparisonReport) {
	fmt.Println("\n🎯 BENCHMARK SUMMARY")
	fmt.Println("==================")
	fmt.Printf("Overall Keyp Performance: %s\n", report.Summary.OverallKeypPerformance)
	fmt.Printf("Average Performance Ratio: %.2fx\n", report.Summary.AvgPerformanceRatio)
	fmt.Printf("Best Keyp Command: %s\n", report.Summary.BestKeypCommand)
	fmt.Printf("Worst Keyp Command: %s\n", report.Summary.WorstKeypCommand)

	fmt.Println("\n📊 COMMAND PERFORMANCE")
	fmt.Println("=====================")
	for _, cmd := range report.Commands {
		status := "🟢"
		if cmd.PerformanceRatio < 0.9 {
			status = "🔴"
		} else if cmd.PerformanceRatio < 1.1 {
			status = "🟡"
		}

		fmt.Printf("%s %s: %.0f ops/sec vs %.0f ops/sec (%.2fx)\n",
			status, cmd.Command, cmd.KeypOpsPerSec, cmd.RedisOpsPerSec, cmd.PerformanceRatio)
	}
}
