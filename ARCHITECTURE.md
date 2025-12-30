# Keyp Architecture Documentation

> **Documentação**: [README.md](README.md) | [TESTS.md](TESTS.md)

## Overview

Keyp implements a **Clean Architecture** design following Uncle Bob's principles and incorporates multiple enterprise patterns from Martin Fowler's catalog. This document provides a comprehensive analysis of the architectural decisions, patterns, and layer responsibilities.

## Architectural Assessment

### Clean Architecture Compliance
- ✅ **Dependency Rule**: Dependencies flow inward correctly
- ✅ **Single Responsibility**: Each layer has clear, focused responsibilities  
- ✅ **Open/Closed**: Extensible without modification
- ✅ **Interface Segregation**: Focused, cohesive interfaces
- ✅ **Dependency Inversion**: Abstractions over concretions

## Layer Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    PRESENTATION LAYER                       │
│                  (Redis Protocol - redcon)                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 APPLICATION LAYER                           │
│                   internal/app/                             │
│              • Server (Application Service)                 │
│              • Connection Management                        │
│              • Protocol Translation                         │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   SERVICE LAYER                             │
│                 internal/service/                           │
│              • Handler (Command Invoker)                    │
│              • Command Registry                             │
│              • Business Logic Coordination                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                   DOMAIN LAYER                              │
│                 internal/domain/                            │
│              • Persister Interface                          │
│              • Domain Types & Constants                     │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│               INFRASTRUCTURE LAYER                          │
│                internal/storage/                            │
│              • LMDB Implementation                          │
│              • Repository Pattern                           │
│              • Resource Management                          │
└─────────────────────────────────────────────────────────────┘
```

## Design Patterns Implementation

### 1. Command Pattern
**Location**: `internal/service/`

```go
type Commands map[string]Command
type Command func(Args) *Result

// Registry-based dispatch
handler.commands = Commands{
    "SET": handler.set,
    "GET": handler.get,
    "DEL": handler.del,
    // ...
}
```

**Benefits**:
- Commands as first-class objects
- Easy extensibility (add new commands without modification)
- Transaction support (MULTI/EXEC pattern)
- Centralized command validation and dispatch

**Implementation**: Textbook implementation with registry pattern integration.

### 2. Repository Pattern
**Location**: `internal/domain/` (interface) + `internal/storage/` (implementation)

```go
// Domain Interface
type Persister interface {
    Set(context.Context, []byte, []byte) error
    Get(context.Context, []byte) ([]byte, error)
    Del(context.Context, ...[]byte) (uint32, error)
    // ...
}

// Infrastructure Implementation
type Client struct {
    env *lmdb.Env
    // ...
}
```

**Benefits**:
- Clean separation between domain and infrastructure
- Testability through interface abstraction
- Storage technology independence
- Proper error translation

### 3. Registry Pattern
**Location**: `internal/service/handler.go`

```go
type Validations map[string]*Validation

handler.validations = Validations{
    "SET": {MinArgs: 3, MaxArgs: 3},
    "GET": {MinArgs: 2, MaxArgs: 2},
    // ...
}
```

**Benefits**:
- Configurable validation rules
- Centralized command metadata
- Easy maintenance and extension

### 4. Object Pool Pattern
**Location**: `internal/service/pool.go`

```go
type Pool struct {
    storage domain.Persister
    // Handler pooling implementation
}
```

**Benefits**:
- Performance optimization through object reuse
- Reduced garbage collection pressure
- Resource management

### 5. Application Service Pattern
**Location**: `internal/app/server.go`

```go
type Server struct {
    storage  domain.Persister
    contexts map[int64]func()
    handlers map[int64]*service.Handler
    poolHdlr *service.Pool
}
```

**Benefits**:
- Orchestrates between layers
- Manages application-level concerns
- Connection lifecycle management

### 6. Gateway Pattern
**Location**: `internal/storage/client.go`

**Benefits**:
- Encapsulates external system (LMDB) complexity
- Provides domain-friendly interface
- Handles resource management

## Layer Responsibilities

### Domain Layer (`internal/domain/`)
**Responsibility**: Core business concepts and contracts

- **Persister Interface**: Defines storage contract
- **Context Constants**: Domain-specific context keys
- **Zero Dependencies**: Pure domain logic

### Service Layer (`internal/service/`)
**Responsibility**: Business logic coordination and command processing

- **Handler**: Command invoker and coordinator
- **Commands**: Individual command implementations
- **Validation**: Business rule enforcement
- **Transaction Management**: MULTI/EXEC support

### Infrastructure Layer (`internal/storage/`)
**Responsibility**: External system integration and persistence

- **LMDB Integration**: Database operations
- **Context Handling**: Cancellation and timeout support
- **TTL Management**: Expiration logic
- **Resource Management**: Connection and cleanup

### Application Layer (`internal/app/`)
**Responsibility**: Application orchestration and protocol handling

- **Server**: Main application coordinator
- **Connection Management**: Per-connection context and handlers
- **Protocol Translation**: Redis protocol to internal commands
- **Lifecycle Management**: Startup and shutdown

## Context-Driven Design

### Context Propagation
```go
// Context carries connection and database information
ctx = context.WithValue(ctx, domain.ID, connID)
ctx = context.WithValue(ctx, domain.DB, dbIndex)
```

**Benefits**:
- Request tracing and cancellation
- Database selection per connection
- Graceful shutdown support

### Cancellation Handling
```go
func ctxFlush(ctx context.Context) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return nil
    }
}
```

**Quality**: - Proper context-aware operations throughout

## Performance Considerations

### Zero-Allocation Goals
- **Buffer Reuse**: Pre-allocated result objects
- **Object Pooling**: Handler reuse across connections
- **Memory-Mapped I/O**: LMDB provides zero-copy reads

### Concurrency Design
- **Per-Connection Handlers**: Isolated command processing
- **Thread-Safe Storage**: Mutex-protected TTL management
- **Context Cancellation**: Graceful operation termination

## Extensibility Points

### Adding New Commands
1. Implement command function: `func (handler *Handler) newCmd(args Args) *Result`
2. Register in Commands map: `"NEWCMD": handler.newCmd`
3. Add validation rules: `"NEWCMD": {MinArgs: X, MaxArgs: Y}`

### Storage Backend Replacement
1. Implement `domain.Persister` interface
2. Replace in dependency injection
3. No other code changes required

### Protocol Extensions
1. Extend `Handler.Apply()` for new protocol features
2. Add new result types if needed
3. Maintain backward compatibility

## Implementation

The architecture provides a solid foundation for both educational purposes and production deployment, demonstrating how to build maintainable, high-performance systems in Go.

---

> 🏠 **Back to Project Overview**: [README.md](README.md)

**Key Strengths**:
- Clean separation of concerns
- Proper dependency management
- Excellent pattern implementation
- Context-aware design
- Performance considerations
