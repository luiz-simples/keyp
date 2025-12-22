#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
BENCHMARK_DIR="$PROJECT_ROOT/benchmarks"
RESULTS_DIR="$BENCHMARK_DIR/results/$(date +%Y-%m-%d)"

echo "🚀 Starting Keyp vs Redis Benchmark"
echo "Results will be saved to: $RESULTS_DIR"

cd "$PROJECT_ROOT"

echo "📦 Building benchmark tool..."
go build -o "$BENCHMARK_DIR/bin/benchmark" "$BENCHMARK_DIR/cmd/benchmark"

echo "🐳 Starting Docker containers..."
docker-compose up -d

echo "⏳ Waiting for services to be ready..."
sleep 10

echo "🔍 Testing connectivity..."
if ! docker exec keyp-redis-test redis-cli ping > /dev/null 2>&1; then
    echo "❌ Redis container not ready"
    exit 1
fi

if ! nc -z localhost 6380 > /dev/null 2>&1; then
    echo "❌ Keyp container not ready"
    exit 1
fi

echo "✅ Both services are ready"

echo "📊 Running Redis benchmark..."
"$BENCHMARK_DIR/bin/benchmark" \
    -server=redis \
    -addr=localhost:6379 \
    -ops=50000 \
    -clients=20 \
    -keysize=16 \
    -valuesize=64 \
    -ttl=300 \
    -output="$RESULTS_DIR/redis"

echo "📊 Running Keyp benchmark..."
"$BENCHMARK_DIR/bin/benchmark" \
    -server=keyp \
    -addr=localhost:6380 \
    -ops=50000 \
    -clients=20 \
    -keysize=16 \
    -valuesize=64 \
    -ttl=300 \
    -output="$RESULTS_DIR/keyp"

echo "📈 Generating comparison report..."
go run "$BENCHMARK_DIR/cmd/compare" \
    -redis="$RESULTS_DIR/redis" \
    -keyp="$RESULTS_DIR/keyp" \
    -output="$RESULTS_DIR/comparison"

echo "🧹 Cleaning up containers..."
docker-compose down

echo "✅ Benchmark completed successfully!"
echo "📁 Results available at: $RESULTS_DIR"