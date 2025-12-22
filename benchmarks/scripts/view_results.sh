#!/bin/bash

BENCHMARK_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
RESULTS_DIR="$BENCHMARK_DIR/results"

if [ $# -eq 0 ]; then
    echo "📊 Available benchmark results:"
    echo "=============================="
    
    if [ ! -d "$RESULTS_DIR" ]; then
        echo "❌ No results directory found. Run a benchmark first."
        exit 1
    fi
    
    for date_dir in "$RESULTS_DIR"/*; do
        if [ -d "$date_dir" ]; then
            date=$(basename "$date_dir")
            echo "📅 $date"
            
            if [ -d "$date_dir/comparison" ]; then
                echo "   📈 Comparison reports available"
                for file in "$date_dir/comparison"/*.md; do
                    if [ -f "$file" ]; then
                        echo "      - $(basename "$file")"
                    fi
                done
            fi
            
            if [ -d "$date_dir/keyp" ]; then
                echo "   🔑 Keyp results available"
            fi
            
            if [ -d "$date_dir/redis" ]; then
                echo "   🔴 Redis results available"
            fi
            echo
        fi
    done
    
    echo "Usage: $0 [date] [type]"
    echo "  date: YYYY-MM-DD (default: latest)"
    echo "  type: comparison|keyp|redis (default: comparison)"
    echo
    echo "Examples:"
    echo "  $0                    # Show this list"
    echo "  $0 latest             # Show latest comparison"
    echo "  $0 2024-01-15         # Show specific date comparison"
    echo "  $0 latest keyp        # Show latest Keyp results"
    
    exit 0
fi

DATE="$1"
TYPE="${2:-comparison}"

if [ "$DATE" = "latest" ]; then
    DATE=$(ls -1 "$RESULTS_DIR" | sort -r | head -1)
fi

RESULT_PATH="$RESULTS_DIR/$DATE/$TYPE"

if [ ! -d "$RESULT_PATH" ]; then
    echo "❌ Results not found: $RESULT_PATH"
    exit 1
fi

echo "📊 Showing $TYPE results for $DATE"
echo "=================================="

case "$TYPE" in
    "comparison")
        for file in "$RESULT_PATH"/*.md; do
            if [ -f "$file" ]; then
                echo "📈 $(basename "$file")"
                echo "---"
                cat "$file"
                echo
            fi
        done
        ;;
    "keyp"|"redis")
        for file in "$RESULT_PATH"/*.json; do
            if [ -f "$file" ]; then
                echo "📊 $(basename "$file")"
                echo "---"
                if command -v jq &> /dev/null; then
                    jq '.' "$file"
                else
                    cat "$file"
                fi
                echo
            fi
        done
        ;;
    *)
        echo "❌ Unknown type: $TYPE"
        echo "Available types: comparison, keyp, redis"
        exit 1
        ;;
esac