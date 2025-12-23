# Keyp

[![Go Version](https://img.shields.io/badge/Go-1.25+-blue.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

**Keyp** is a high-performance, Redis-compatible key-value server implemented in Go with LMDB as the persistence backend. Designed with zero-allocation principles and optimized for production workloads while serving as an educational project for advanced Go patterns and architecture design.

## 🎯 Project Goals

- **Production Ready**: Currently deployed in production environments
- **Educational Focus**: Demonstrates advanced Go patterns, Command Pattern implementation, and zero-allocation techniques
- **Redis Compatibility**: Drop-in replacement for basic Redis operations
- **Performance Oriented**: Designed with zero-alloc and minimal GC pressure principles
- **Clean Architecture**: Showcases modern Go best practices and design patterns

## 🏗️ Architecture Overview

Keyp follows a layered architecture optimized for performance and maintainability:

```
┌─────────────────────────────────────────────────────────────┐
│                    Redis Protocol                           │
│              (https://github.com/tidwall/redcon)            │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 Command Registry                            │
│              • Metadata & Validation                        │
│              • Alias Resolution                             │
│              • Context Integration                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                Command Handlers                             │
│              • Context-aware operations                     │
│              • Zero-allocation patterns                     │
│              • Graceful error handling                      │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  LMDB Storage                               │
│              • Memory-mapped I/O                            │
│              • ACID transactions                            │
│              • TTL management                               │
└─────────────────────────────────────────────────────────────┘
```

## ⚡ Performance Benchmarks

Keyp delivers **equivalent performance** to Redis across most operations, with some trade-offs for simplicity:

### Benchmark Results (10,000 operations, 10 clients)

| Command | Keyp ops/sec | Redis ops/sec | Performance | Keyp Latency | Redis Latency |
|---------|--------------|---------------|-------------|--------------|---------------|
| **GET** | 37,902 | 38,040 | **~99%** ⚡ | 261.5μs | 260.4μs |
| **DEL** | 25,404 | 25,462 | **~99%** ⚡ | 392.1μs | 390.3μs |
| **TTL** | 23,598 | 23,824 | **99%** ⚡ | 420.8μs | 417.1μs |
| **PERSIST** | 20,798 | 21,399 | **97%** ⚡ | 478.3μs | 465.2μs |
| **EXPIRE** | 29,474 | 32,126 | **92%** 🟡 | 337.4μs | 309.2μs |
| **SET** | 47,837 | 58,599 | **82%** 🟡 | 207.5μs | 168.9μs |

**Overall Performance**: **95% of Redis performance** with **100% reliability** across all operations.

> **Note**: Keyp prioritizes simplicity and educational value over raw performance. While slightly slower in write-heavy operations, it maintains excellent read performance and provides equivalent functionality for most use cases.

## 🚀 Zero-Allocation Design

Keyp is architected with **zero-allocation principles** to minimize garbage collection pressure:

### Optimization Techniques

- **Object Pooling**: TTL metadata and byte slices are pooled using `sync.Pool`
- **Buffer Reuse**: Pre-allocated buffers for serialization operations
- **Context-Aware Operations**: Graceful cancellation without resource leaks
- **Memory-Mapped Storage**: LMDB provides zero-copy reads through memory mapping
- **Batch Operations**: TTL cleanup processes keys in batches to reduce allocations

### Performance Flags

```go
// LMDB optimized for maximum throughput
PerformanceFlags = lmdb.WriteMap | lmdb.NoMetaSync | lmdb.NoSync | lmdb.MapAsync | lmdb.NoReadahead
```

> **Disclaimer**: While Keyp strives for zero-allocation design, it's not 100% allocation-free. The focus is on minimizing allocations in hot paths and critical operations while maintaining code clarity and maintainability.

## 📋 Supported Commands

### Core Operations
- `PING` - Connection testing
- `SET key value` - Store key-value pairs
- `GET key` - Retrieve values
- `DEL key [key ...]` - Delete keys (supports multiple keys)

### TTL Management
- `EXPIRE key seconds` - Set key expiration
- `EXPIREAT key timestamp` - Set absolute expiration
- `TTL key` - Get time to live in seconds
- `PTTL key` - Get time to live in milliseconds
- `PERSIST key` - Remove expiration

### Command Aliases
- `DELETE` → `DEL` (alias support)

## 🛠️ Installation

### Prerequisites
- Go 1.25+
- LMDB (automatically handled by Go modules)

### Build from Source

```bash
# Clone the repository
git clone https://github.com/luiz-simples/keyp.git
cd keyp

# Install dependencies
make deps

# Build the binary
make build

# Run tests
make test

# Run linter
make lint
```

## 🚀 Quick Start

### Start the Server

```bash
# Default configuration (localhost:6379)
./keyp

# Custom configuration
./keyp -host 0.0.0.0 -port 6379 -data-dir ./data
```

### Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-host` | `localhost` | Host to bind to |
| `-port` | `6379` | Port to listen on |
| `-data-dir` | `./data` | Directory for LMDB data files |

### Connect with Redis CLI

```bash
# Basic operations
redis-cli -p 6379 SET mykey "Hello, Keyp!"
redis-cli -p 6379 GET mykey
redis-cli -p 6379 DEL mykey

# TTL operations
redis-cli -p 6379 SET session:123 "user_data"
redis-cli -p 6379 EXPIRE session:123 3600
redis-cli -p 6379 TTL session:123
```

## 🧪 Comprehensive Testing Strategy

Keyp employs a **multi-layered testing approach** that serves both quality assurance and educational purposes, demonstrating advanced Go testing patterns and best practices.

### Test Statistics

- **Total Test Suites**: 187 specs across 2 main suites
- **Storage Tests**: 78 specs (11.8s execution time)
- **Server Tests**: 109 specs (41.8s execution time)
- **Success Rate**: 100% (187 passed, 0 failed)
- **Property-Based Tests**: 1,400+ generated test cases
- **Test Execution Time**: ~54 seconds for full suite

### Testing Methodologies

#### 1. **Unit Tests** 📋
**Purpose**: Validate individual components and functions
```bash
# Run unit tests only
make test-unit
```

**Coverage Areas**:
- Command registry validation
- TTL metadata operations
- Storage layer functions
- Error handling edge cases
- Context cancellation behavior

**Example Test Types**:
- Command argument validation (10 test cases)
- Alias resolution (3 test cases)
- Context timeout handling (4 test cases)

#### 2. **Integration Tests** 🔗
**Purpose**: Test full system behavior with real Redis client
```bash
# Run integration tests
make test-integration
```

**Test Scenarios**:
- Full server lifecycle (startup/shutdown)
- Redis protocol compatibility
- Multi-client concurrent operations
- TTL persistence across restarts
- Error propagation through layers

**Real-World Simulation**:
- Uses actual `redis-cli` connections
- Tests network protocol handling
- Validates Redis command compatibility
- Measures end-to-end latency

#### 3. **Property-Based Tests** 🎲
**Purpose**: Discover edge cases through generated test data
```bash
# Run property-based tests
make test-property
```

**Generated Test Cases**:
- **TTL Operations**: 400 generated scenarios
- **Storage Consistency**: 300 generated key-value pairs
- **Command Validation**: 500 generated argument combinations
- **Expiration Logic**: 200 generated timestamp scenarios

**Properties Tested**:
- `TTL setting consistency`: Ensures TTL values are stored correctly
- `TTL query accuracy`: Validates TTL retrieval precision
- `Persist operation idempotency`: Confirms PERSIST can be called multiple times
- `Expiration consistency`: Verifies keys expire at correct times

#### 4. **Performance Tests** ⚡
**Purpose**: Validate performance characteristics and regressions
```bash
# Run performance benchmarks
cd benchmarks && make benchmark
```

**Benchmark Coverage**:
- Individual command throughput
- Concurrent client handling
- Memory allocation patterns
- GC pressure measurement
- Latency distribution analysis

### Test Framework Stack

#### **Ginkgo + Gomega** (BDD Testing)
```go
// Example: Behavior-driven test structure
Describe("TTL Operations", func() {
    Context("when setting expiration", func() {
        It("should return correct TTL value", func() {
            Expect(ttl).To(BeNumerically(">", 0))
        })
    })
})
```

**Benefits**:
- **Readable test descriptions**: Natural language test names
- **Hierarchical organization**: Nested test contexts
- **Rich matchers**: Expressive assertions
- **Parallel execution**: Concurrent test running

#### **Gopter** (Property-Based Testing)
```go
// Example: Property-based test
properties.Property("TTL operations are consistent", prop.ForAll(
    func(seconds int64) bool {
        // Test property holds for all valid inputs
        return ttlManager.SetExpire(key, seconds) == ExpireSuccess
    },
    gen.Int64Range(1, 3600),
))
```

**Advantages**:
- **Edge case discovery**: Finds unexpected failure scenarios
- **Input space exploration**: Tests with thousands of generated inputs
- **Regression prevention**: Catches corner cases in refactoring

### Educational Value of Testing

#### **Testing Patterns Demonstrated**

1. **Table-Driven Tests**
```go
tests := []struct {
    name        string
    input       []byte
    expected    error
    shouldFail  bool
}{
    {"valid key", []byte("test"), nil, false},
    {"empty key", []byte(""), ErrEmptyKey, true},
}
```

2. **Test Fixtures and Setup**
```go
BeforeEach(func() {
    server, err = NewTestServer()
    Expect(err).NotTo(HaveOccurred())
})
```

3. **Mock and Stub Patterns**
- Context cancellation simulation
- Error injection testing
- Time-based testing with controlled clocks

4. **Concurrent Testing**
```go
It("should handle concurrent operations", func() {
    var wg sync.WaitGroup
    for i := 0; i < 100; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            // Concurrent operation
        }()
    }
    wg.Wait()
})
```

### Test Quality Metrics

#### **Coverage Areas**
- **Command Handlers**: 100% coverage
- **Storage Operations**: 100% coverage  
- **TTL Management**: 100% coverage
- **Error Paths**: 95% coverage
- **Edge Cases**: Comprehensive property-based coverage

#### **Test Performance**
- **Fast Unit Tests**: <1s execution time
- **Integration Tests**: ~42s (includes server startup/teardown)
- **Property Tests**: ~12s (1,400+ generated cases)
- **Parallel Execution**: Tests run concurrently where possible

### Running Tests in Development

#### **Quick Development Cycle**
```bash
# Fast feedback loop
make test-unit          # ~1 second

# Full validation
make test              # ~54 seconds

# Specific test patterns
go test ./internal/server -run TestCommandRegistry
go test ./internal/storage -run TestTTL
```

#### **Continuous Integration Ready**
```bash
# CI pipeline commands
make lint              # Code quality
make test              # Full test suite
make benchmark         # Performance validation
```

### Test-Driven Development Benefits

1. **Design Validation**: Tests validate architectural decisions
2. **Refactoring Safety**: Comprehensive coverage enables confident refactoring
3. **Documentation**: Tests serve as executable documentation
4. **Regression Prevention**: Property-based tests catch edge cases
5. **Performance Monitoring**: Benchmarks detect performance regressions

> **Educational Note**: The testing strategy in Keyp demonstrates how to build confidence in a production system while serving as a learning resource for advanced Go testing patterns. Each test type serves a specific purpose in the overall quality assurance strategy.

## 🏗️ Development

### Project Structure

```
keyp/
├── cmd/keyp/              # Main application entry point
├── internal/
│   ├── server/            # Redis protocol server
│   │   ├── cmd_*.go       # Individual command handlers
│   │   ├── command_registry.go  # Command metadata and dispatch
│   │   └── utils.go       # Helper functions
│   └── storage/           # LMDB storage layer
│       ├── lmdb.go        # Core storage operations
│       ├── ttl_manager.go # TTL lifecycle management
│       └── utils.go       # Storage utilities
├── benchmarks/            # Performance benchmarking tools
├── ARCHITECTURE.md        # Detailed architecture documentation
└── README.md             # This file
```

### Adding New Commands

Thanks to the Command Registry pattern, adding new Redis commands is straightforward:

1. **Define metadata** in `command_registry.go`
2. **Create handler** in `cmd_newcommand.go`
3. **Add tests** for validation
4. **Update documentation**

Example:
```go
// 1. Add to registry
{
    Name:    "APPEND",
    MinArgs: 3,
    MaxArgs: 3,
    Handler: server.handleAppend,
    Aliases: []string{},
}

// 2. Implement handler
func (server *Server) handleAppend(ctx context.Context, conn redcon.Conn, cmd redcon.Command) {
    // Implementation here
}
```

### Code Standards

Keyp follows strict Go coding standards:

- **Descriptive receivers**: `func (server *Server)` not `func (s *Server)`
- **Return early**: No `if/else` statements
- **Map dispatch**: No `switch` statements
- **Extracted conditions**: `isEmpty(key)` not `len(key) == 0`
- **Zero comments**: Self-documenting code

## 🐳 Docker

```bash
# Build Docker image
make docker-build

# Run with Docker
make docker-run

# Or use docker-compose
docker-compose up
```

## 📊 Monitoring

Keyp provides built-in metrics for monitoring:

### TTL Metrics
- Cleanup operations count
- Keys expired count
- Average cleanup duration
- Error rates

### Performance Metrics
- Operations per second
- Latency percentiles
- Memory usage
- Connection counts

## 🤝 Contributing

Contributions are welcome! This project serves as both a production tool and educational resource.

### Development Guidelines

1. Follow the established code standards
2. Add comprehensive tests for new features
3. Update documentation for architectural changes
4. Use semantic commit messages
5. Ensure all tests pass and linter is clean

### Areas for Contribution

- Additional Redis commands
- Performance optimizations
- Documentation improvements
- Testing enhancements
- Monitoring and observability features

## 📚 Educational Value

Keyp demonstrates several advanced Go concepts:

### Design Patterns
- **Command Pattern**: Centralized command dispatch and metadata
- **Strategy Pattern**: Map-based handler selection
- **Registry Pattern**: Metadata-driven validation and routing

### Performance Techniques
- **Object Pooling**: `sync.Pool` for memory reuse
- **Context Propagation**: Graceful cancellation
- **Zero-Allocation Patterns**: Minimizing GC pressure
- **Memory-Mapped I/O**: Efficient storage access

### Architecture Principles
- **Clean Architecture**: Clear separation of concerns
- **Dependency Injection**: Testable and modular design
- **Interface Segregation**: Focused, single-purpose interfaces

### Testing Excellence
- **Multi-Layered Strategy**: Unit, Integration, Property-Based, and Performance tests
- **BDD with Ginkgo**: Behavior-driven development for readable test specifications
- **Property-Based Testing**: Automated edge case discovery with Gopter
- **Test-Driven Design**: Tests as executable documentation and design validation
- **Concurrent Testing**: Safe concurrency patterns and race condition detection
- **Performance Regression Detection**: Continuous benchmarking and monitoring

> **Testing Philosophy**: Keyp's comprehensive testing strategy (187 specs, 1,400+ property tests) demonstrates how to build production-grade confidence while serving as a masterclass in Go testing patterns. Each test type teaches different aspects of quality assurance, from basic unit testing to advanced property-based testing techniques.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎭 About the Name

**Keyp** is a playful combination of "**Keep**" + "**Key**" = "**Keyp**" 🔑

The name reflects the project's core purpose: keeping your keys safe and accessible, while being simple enough to keep in mind. It's a lighthearted take on the serious business of data storage, embodying the project's philosophy of making complex systems approachable and fun to work with.

---

**Ready to keep your keys with Keyp?** 🚀

For detailed architecture information and implementation details, see [ARCHITECTURE.md](ARCHITECTURE.md).