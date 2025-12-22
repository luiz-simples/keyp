#!/bin/bash

set -e

echo "📦 Installing benchmark dependencies..."

if ! command -v docker &> /dev/null; then
    echo "❌ Docker is required but not installed"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "❌ Docker Compose is required but not installed"
    exit 1
fi

if ! command -v go &> /dev/null; then
    echo "❌ Go is required but not installed"
    exit 1
fi

if ! command -v nc &> /dev/null; then
    echo "⚠️  netcat (nc) not found, installing..."
    if [[ "$OSTYPE" == "darwin"* ]]; then
        if command -v brew &> /dev/null; then
            brew install netcat
        else
            echo "❌ Please install netcat manually"
            exit 1
        fi
    elif [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if command -v apt-get &> /dev/null; then
            sudo apt-get update && sudo apt-get install -y netcat
        elif command -v yum &> /dev/null; then
            sudo yum install -y nc
        else
            echo "❌ Please install netcat manually"
            exit 1
        fi
    fi
fi

echo "✅ All dependencies are installed"

echo "🔧 Setting up Go modules..."
go mod tidy

echo "🐳 Pulling Docker images..."
docker pull redis:alpine
docker pull golang:1.21-alpine

echo "✅ Setup completed successfully!"