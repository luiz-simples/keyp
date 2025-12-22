.PHONY: benchmark benchmark-keyp benchmark-redis benchmark-report clean setup

BENCHMARK_DIR := benchmarks
BIN_DIR := $(BENCHMARK_DIR)/bin
RESULTS_DIR := $(BENCHMARK_DIR)/results/$(shell date +%Y-%m-%d)

setup:
	@echo "🔧 Setting up benchmark environment..."
	@mkdir -p $(BIN_DIR)
	@mkdir -p $(RESULTS_DIR)
	@go mod tidy

build-tools: setup
	@echo "🔨 Building benchmark tools..."
	@go build -o $(BIN_DIR)/benchmark $(BENCHMARK_DIR)/cmd/benchmark
	@go build -o $(BIN_DIR)/compare $(BENCHMARK_DIR)/cmd/compare

benchmark: build-tools
	@echo "🚀 Running complete benchmark suite..."
	@./$(BENCHMARK_DIR)/scripts/run_benchmark.sh

benchmark-keyp: build-tools
	@echo "📊 Running Keyp benchmark only..."
	@docker-compose up -d keyp
	@sleep 5
	@$(BIN_DIR)/benchmark \
		-server=keyp \
		-addr=localhost:6380 \
		-ops=50000 \
		-clients=20 \
		-keysize=16 \
		-valuesize=64 \
		-ttl=300 \
		-output="$(RESULTS_DIR)/keyp"
	@docker-compose down

benchmark-redis: build-tools
	@echo "📊 Running Redis benchmark only..."
	@docker-compose up -d redis
	@sleep 5
	@$(BIN_DIR)/benchmark \
		-server=redis \
		-addr=localhost:6379 \
		-ops=50000 \
		-clients=20 \
		-keysize=16 \
		-valuesize=64 \
		-ttl=300 \
		-output="$(RESULTS_DIR)/redis"
	@docker-compose down

benchmark-report:
	@echo "📈 Generating comparison report..."
	@$(BIN_DIR)/compare \
		-redis="$(RESULTS_DIR)/redis" \
		-keyp="$(RESULTS_DIR)/keyp" \
		-output="$(RESULTS_DIR)/comparison"

benchmark-quick: build-tools
	@echo "⚡ Running quick benchmark (10k ops, 5 clients)..."
	@docker-compose up -d
	@sleep 10
	@$(BIN_DIR)/benchmark \
		-server=redis \
		-addr=localhost:6379 \
		-ops=10000 \
		-clients=5 \
		-keysize=16 \
		-valuesize=64 \
		-ttl=300 \
		-output="$(RESULTS_DIR)/redis"
	@$(BIN_DIR)/benchmark \
		-server=keyp \
		-addr=localhost:6380 \
		-ops=10000 \
		-clients=5 \
		-keysize=16 \
		-valuesize=64 \
		-ttl=300 \
		-output="$(RESULTS_DIR)/keyp"
	@$(BIN_DIR)/compare \
		-redis="$(RESULTS_DIR)/redis" \
		-keyp="$(RESULTS_DIR)/keyp" \
		-output="$(RESULTS_DIR)/comparison"
	@docker-compose down

clean:
	@echo "🧹 Cleaning up..."
	@docker-compose down -v
	@rm -rf $(BIN_DIR)
	@docker system prune -f

help:
	@echo "Available targets:"
	@echo "  benchmark       - Run complete benchmark suite (50k ops, 20 clients)"
	@echo "  benchmark-quick - Run quick benchmark (10k ops, 5 clients)"
	@echo "  benchmark-keyp  - Run Keyp benchmark only"
	@echo "  benchmark-redis - Run Redis benchmark only"
	@echo "  benchmark-report- Generate comparison report from existing results"
	@echo "  setup          - Setup benchmark environment"
	@echo "  build-tools    - Build benchmark tools"
	@echo "  clean          - Clean up containers and build artifacts"
	@echo "  help           - Show this help message"