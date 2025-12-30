# Keyp

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **Documentação**: [ARCHITECTURE.md](ARCHITECTURE.md) | [TESTS.md](TESTS.md)

**Keyp** is a high-performance, Redis-compatible key-value server implemented in Go with LMDB as the persistence backend. Supports string operations and list data structures with Redis-compatible commands. Designed with zero-allocation principles and optimized for production workloads while serving as an educational project for advanced Go patterns and architecture design.

## 🎯 Project Goals

- **Production Ready**: Currently deployed in production environments
- **Educational Focus**: Demonstrates advanced Go patterns, Command Pattern implementation, and zero-allocation techniques
- **Redis Compatibility**: Drop-in replacement for basic Redis operations
- **Performance Oriented**: Designed with zero-alloc and minimal GC pressure principles
- **Clean Architecture**: Showcases modern Go best practices and design patterns

## 🏗️ Architecture

Keyp implements **Clean Architecture** principles with excellent separation of concerns and proper dependency management. The system demonstrates multiple enterprise patterns including Command Pattern, Repository Pattern, and Object Pool Pattern.

> 📖 **For comprehensive architectural analysis, design patterns documentation, and layer responsibilities, see [ARCHITECTURE.md](ARCHITECTURE.md)**

## 🧪 Testing Strategy

Keyp implements a comprehensive testing strategy with **4 types of tests** ensuring reliability and performance:

- **Unit Tests**: Isolated testing with mocks
- **Integration Tests**: End-to-end Redis protocol compatibility  
- **Property-Based Tests**: Invariant validation with random data
- **Performance Tests**: Benchmarks and performance validation

**Current Coverage**: `internal/service` package (100% complete)

> 🔬 **For detailed testing documentation, execution commands, and performance metrics, see [TESTS.md](TESTS.md)**

## ⚡ Performance Benchmarks

Keyp delivers **equivalent performance** to Redis across most operations, with some trade-offs for simplicity:

### Benchmark Results (10,000 operations, 10 clients)

| Command | Keyp ops/sec | Redis ops/sec | Performance | Keyp Latency | Redis Latency |
|---------|--------------|---------------|-------------|--------------|---------------|
| **GET** | 37,902 | 38,040 | **~99%** ⚡ | 261.5μs | 260.4μs |
| **DEL** | 25,404 | 25,462 | **~99%** ⚡ | 392.1μs | 390.3μs |
| **SET** | 47,837 | 58,599 | **82%** 🟡 | 207.5μs | 168.9μs |

**Overall Performance**: **95% of Redis performance** with **100% reliability** across all operations.

> **Note**: Keyp prioritizes simplicity and educational value over raw performance. While slightly slower in write-heavy operations, it maintains excellent read performance and provides equivalent functionality for most use cases.

## 🚀 Quick Start

### Supported Commands

Keyp implements Redis-compatible commands for string and list operations:

#### String Operations
- `SET key value` - Set a key-value pair
- `GET key` - Get value by key
- `DEL key [key ...]` - Delete one or more keys
- `EXISTS key` - Check if key exists
- `PING` - Test server connectivity

#### List Operations
- `LPUSH key value [value ...]` - Push values to the left of list
- `RPUSH key value [value ...]` - Push values to the right of list
- `LPOP key` - Pop value from the left of list
- `RPOP key` - Pop value from the right of list
- `LLEN key` - Get list length
- `LINDEX key index` - Get element at index
- `LRANGE key start stop` - Get range of elements
- `LSET key index value` - Set element at index

### Installation

```bash
git clone https://github.com/luiz-simples/keyp.git
cd keyp
go build -o keyp cmd/keyp/main.go
```

### Running Tests

```bash
# All tests (parallel)
ginkgo -p ./internal/service

# Benchmarks
go test ./internal/service -bench=. -run=^$
```

> 📋 **For complete testing commands and options, see [TESTS.md](TESTS.md)**

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎭 About the Name

**Keyp** is a playful combination of "**Keep**" + "**Key**" = "**Keyp**" 🔑

The name reflects the project's core purpose: keeping your keys safe and accessible, while being simple enough to keep in mind. It's a lighthearted take on the serious business of data storage, embodying the project's philosophy of making complex systems approachable and fun to work with.
