# Keyp - Testing Strategy

> **Related Documentation**: [README.md](README.md) | [ARCHITECTURE.md](ARCHITECTURE.md)

## Overview

This document describes the complete testing strategy for the Keyp project, implemented following the code standards defined in the project.

## Current Coverage

### ✅ Package `internal/service` (Complete)

Complete implementation of 4 test types with total isolation and safe parallelization.

#### 🧪 Unit Tests
- **File**: `internal/service/unit_test.go`
- **Framework**: Ginkgo + Gomega + GoMock
- **Coverage**: Basic commands (PING, SET, GET, DEL), list commands (EXISTS, LINDEX, LLEN, LPOP, LPUSH, LRANGE, LSET, RPOP, RPUSH), set commands (FLUSHALL, SADD, SREM, SMEMBERS, SISMEMBER), sorted set commands (ZADD, ZRANGE, ZCOUNT), numeric commands (INCR, INCRBY, DECR, DECRBY) and string command (APPEND)
- **Scenarios**: Success, error, canceled context, validation, list operations
- **Mocks**: Generated with `mockgen` for `domain.Persister`

#### 🔗 Integration Tests  
- **File**: `internal/service/integration_test.go`
- **Framework**: Ginkgo + Gomega + go-redis
- **Coverage**: Real Redis server with go-redis client, basic, list, set, sorted set, numeric and string commands
- **Scenarios**: Basic operations, list operations, concurrency, large values
- **Protocol**: Full Redis compatibility

#### 🎯 Property-Based Tests
- **File**: `internal/service/property_test.go` 
- **Framework**: Ginkgo + Gomega + Gopter
- **Coverage**: Fundamental properties (SET-GET, DEL, EXISTS), list operations, set operations, sorted set operations and numeric operations
- **Scenarios**: 100 tests per property with random data
- **Validation**: Invariants and expected behaviors for strings and lists

#### ⚡ Performance Tests
- **File**: `internal/service/performance_test.go`
- **Framework**: Ginkgo + Native Go Benchmarks
- **Coverage**: Performance metrics for basic and list commands
- **Scenarios**: SET, GET, DEL, PING, EXISTS, list operations (LPUSH, RPUSH, etc.)
- **Benchmarks**: Precise ns/op metrics

### ✅ Package `internal/storage` (Complete)

Complete implementation of 4 test types for the LMDB persistence system with total isolation.

#### 🧪 Unit Tests (26 tests)
- **File**: `internal/storage/unit_test.go`
- **Framework**: Ginkgo + Gomega
- **Coverage**: All LMDB operations (Set, Get, Del, TTL, Expire, Persist, EXISTS, list operations)
- **Scenarios**: Client creation, database isolation, error handling
- **Validation**: Empty keys, large values, canceled contexts, list operations

#### 🔗 Integration Tests (12 tests)
- **File**: `internal/storage/integration_test.go`
- **Framework**: Ginkgo + Gomega
- **Coverage**: Real LMDB instances with concurrent operations
- **Scenarios**: Multiple goroutines, database isolation, large data
- **Validation**: Thread-safety, TTL, timeout and context cancellation

#### 🎯 Property-Based Tests (10 tests)
- **File**: `internal/storage/property_test.go`
- **Framework**: Ginkgo + Gomega + Gopter
- **Coverage**: Storage invariants (Set-Get, Set-Delete, TTL, Persist)
- **Scenarios**: 1000 tests per property (100 executions × 10 properties)
- **Validation**: Database isolation, idempotency, consistency

#### ⚡ Performance Tests (12 tests + benchmarks)
- **File**: `internal/storage/performance_test.go`
- **Framework**: Ginkgo + gmeasure + Go Benchmarks
- **Coverage**: Individual and batch performance, concurrency, throughput
- **Scenarios**: Individual operations, batch operations, large data
- **Benchmarks**: Precise ns/op metrics for LMDB

### 🔄 Next Packages

- `internal/app` - Planned  
- `cmd/keyp` - Planned

## Test Architecture

### Isolation and Parallelization

#### ✅ Implemented Features

- **Unique Directories**: Each test uses a unique temporary directory
- **Automatic Cleanup**: Directories are removed after each test
- **Safe Parallelization**: Tests can run in parallel without conflicts
- **Total Isolation**: No test interferes with another

#### 📁 Directory Pattern

```
/tmp/keyp-{type}-{pid}-{timestamp}/
```

Examples:
- `/tmp/keyp-integration-12345-1766756124000/`
- `/tmp/keyp-property-12345-1766756125000/`
- `/tmp/keyp-bench-set-12345-1766756126000/`

#### 🔄 Automatic Cleanup

- **BeforeEach**: Creates unique directory
- **AfterEach**: Removes directory and closes storage
- **Benchmarks**: Uses `defer` for guaranteed cleanup

## Test Execution

### Main Commands

#### All Tests (Parallel - Recommended)
```bash
# Service package
ginkgo -p ./internal/service

# Storage package  
ginkgo -p ./internal/storage

# All packages
ginkgo -p ./internal/...
```

#### All Tests (Sequential)
```bash
# Service package
go test ./internal/service -v

# Storage package
go test ./internal/storage -v

# All packages
go test ./internal/... -v
```

#### By Test Type
```bash
# Unit Tests
go test ./internal/service -v --ginkgo.label-filter="unit"
go test ./internal/storage -v --ginkgo.label-filter="unit"

# Integration Tests  
go test ./internal/service -v --ginkgo.label-filter="integration"
go test ./internal/storage -v --ginkgo.label-filter="integration"

# Property Tests
go test ./internal/service -v --ginkgo.label-filter="property"
go test ./internal/storage -v --ginkgo.label-filter="property"

# Performance Tests
go test ./internal/service -v --ginkgo.label-filter="performance"
go test ./internal/storage -v --ginkgo.label-filter="performance"
```

#### Benchmarks
```bash
# Service benchmarks
go test ./internal/service -bench=. -run=^$

# Storage benchmarks
go test ./internal/storage -bench=. -run=^$

# Specific benchmark
go test ./internal/service -bench=BenchmarkHandlerSET -run=^$
go test ./internal/storage -bench=BenchmarkStorageSet -run=^$
```

### Performance Results

#### Service Package (Apple M1 Pro)

```
BenchmarkHandlerSET-10      	  917700	      1296 ns/op	     120 B/op	       3 allocs/op
BenchmarkHandlerGET-10      	  965050	      1212 ns/op	     136 B/op	       4 allocs/op
BenchmarkHandlerDEL-10      	  390798	      3069 ns/op	     440 B/op	      14 allocs/op
BenchmarkHandlerPING-10     	19328184	        61.07 ns/op	      56 B/op	       2 allocs/op
BenchmarkHandlerMixed-10    	  275817	      4372 ns/op	     624 B/op	      20 allocs/op
```

#### Storage Package (Apple M1 Pro)

```
BenchmarkStorageSet-10      	 1015824	      1186 ns/op	      64 B/op	       1 allocs/op
BenchmarkStorageGet-10      	 1000000	      1112 ns/op	      80 B/op	       2 allocs/op
BenchmarkStorageDel-10      	  432428	      2842 ns/op	     200 B/op	       5 allocs/op
BenchmarkStorageMixed-10    	  298111	      4074 ns/op	     311 B/op	       9 allocs/op
```

#### Performance Validations

**Service Layer:**
- **SET**: < 1 second for 1000 operations
- **GET**: < 1 second for 1000 operations  
- **PING**: < 0.5 seconds for 10000 operations
- **Mixed**: < 3 seconds for 1000 complete operations

**Storage Layer:**
- **SET**: < 1 second for 1000 LMDB operations
- **GET**: < 1 second for 1000 LMDB operations
- **DEL**: < 3 seconds for 1000 LMDB operations
- **Mixed**: < 4 seconds for 1000 complete operations

#### Parallelization

**Service Package:**
```
Sequential: 4.2s (96 specs)
Parallel:   2.1s (96 specs) - 50% faster
Processes:  10 parallel
```

**Storage Package:**
```
Sequential: 7.6s (60 specs)
Parallel:   4.2s (60 specs) - 45% faster  
Processes:  10 parallel
```

## Code Standards in Tests

### Steering Rules Compliance

Tests strictly follow the standards defined in `.kiro/steering/code-standards.md`:

- ✅ **Zero comments** - Descriptive names
- ✅ **Return early** - No `if/else`
- ✅ **Extracted functions** - `hasError()`, `isEmpty()`
- ✅ **Descriptive receivers** - `handler`, not `h`
- ✅ **Semantic commits** - `test: add comprehensive coverage`

### File Structure

**Service Package:**
```
internal/service/
├── service_test.go      # Main suite + utilities
├── mocks_test.go        # Generated mocks (mockgen)
├── unit_test.go         # Unit tests
├── integration_test.go  # Integration tests
├── property_test.go     # Property-based tests
└── performance_test.go  # Performance tests
```

**Storage Package:**
```
internal/storage/
├── storage_test.go      # Main suite + utilities
├── unit_test.go         # Unit tests (26 specs)
├── integration_test.go  # Integration tests (12 specs)
├── property_test.go     # Property-based tests (10 specs)
└── performance_test.go  # Performance tests (12 specs + benchmarks)
```

### Centralized Utilities

```go
func createUniqueTestDir(prefix string) string
func cleanupTestDir(dir string)
```

## Test Dependencies

### Main Frameworks

- `github.com/onsi/ginkgo/v2` - BDD testing framework
- `github.com/onsi/gomega` - Matchers for assertions
- `go.uber.org/mock` - Mock generation
- `github.com/leanovate/gopter` - Property-based testing

### Integration Dependencies

- `github.com/redis/go-redis/v9` - Redis client
- `github.com/tidwall/redcon` - Redis-compatible server

### Mock Generation

```bash
mockgen -source=internal/domain/types.go -destination=internal/service/mocks_test.go -package=service_test
```

## CI/CD Integration

### Recommended Commands

```bash
# Quick check - all packages
go test ./internal/... -v

# Complete check with parallelization
ginkgo -p ./internal/...

# Benchmarks for metrics
go test ./internal/service -bench=. -run=^$ -benchmem
go test ./internal/storage -bench=. -run=^$ -benchmem

# By specific package
go test ./internal/service -v --ginkgo.v
go test ./internal/storage -v --ginkgo.v
```

### Quality Metrics

**Service Package:**
- **Coverage**: 100% of main commands (25 total commands)
- **Specs**: 96 tests in 4 types
- **Isolation**: Total between executions
- **Performance**: Automated benchmarks

**Storage Package:**
- **Coverage**: 100% of LMDB operations
- **Specs**: 60 tests in 4 types
- **Isolation**: Total between executions
- **Performance**: Automated benchmarks

**Project Total:**
- **Specs**: 156 tests (96 service + 60 storage)
- **Property Tests**: 1600+ executions (100 × 16 properties)
- **Benchmarks**: 9 different benchmarks
- **Coverage**: 2 complete packages

---

> **Next Steps**: See [ARCHITECTURE.md](ARCHITECTURE.md) to understand the system structure and [README.md](README.md) for usage instructions.