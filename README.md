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
│                 (github.com/tidwall/redcon)                 │
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
git clone https://github.com/your-username/keyp.git
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
# Default configuration (localhost:6377)
./keyp

# Custom configuration
./keyp -host 0.0.0.0 -port 6377 -data-dir ./data
```

### Command Line Options

| Flag | Default | Description |
|------|---------|-------------|
| `-host` | `localhost` | Host to bind to |
| `-port` | `6377` | Port to listen on |
| `-data-dir` | `./data` | Directory for LMDB data files |

### Connect with Redis CLI

```bash
# Basic operations
redis-cli -p 6377 SET mykey "Hello, Keyp!"
redis-cli -p 6377 GET mykey
redis-cli -p 6377 DEL mykey

# TTL operations
redis-cli -p 6377 SET session:123 "user_data"
redis-cli -p 6377 EXPIRE session:123 3600
redis-cli -p 6377 TTL session:123
```

## 🧪 Testing

Keyp includes comprehensive testing with multiple strategies:

### Test Types

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# Property-based tests
make test-property

# All tests
make test
```

### Test Coverage

- **Unit Tests**: Core functionality and edge cases
- **Integration Tests**: Full server with Redis client
- **Property-Based Tests**: TTL correctness and consistency
- **Performance Tests**: Benchmarks and load testing

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

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🎭 About the Name

**Keyp** is a playful combination of "**Keep**" + "**Key**" = "**Keyp**" 🔑

The name reflects the project's core purpose: keeping your keys safe and accessible, while being simple enough to keep in mind. It's a lighthearted take on the serious business of data storage, embodying the project's philosophy of making complex systems approachable and fun to work with.

---

**Ready to keep your keys with Keyp?** 🚀

For detailed architecture information and implementation details, see [ARCHITECTURE.md](ARCHITECTURE.md).