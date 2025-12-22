# Keyp Architecture Documentation

## Overview

Keyp is a Redis-compatible key-value server implemented in Go with LMDB as the persistence backend. The architecture follows clean separation of concerns with a layered design optimized for performance and maintainability.

## System Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Client Applications                       │
└─────────────────────┬───────────────────────────────────────┘
                      │ Redis Protocol (TCP)
┌─────────────────────▼───────────────────────────────────────┐
│                 Redcon Protocol Layer                       │
│                 (github.com/tidwall/redcon)                 │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Command Registry                           │
│              • Metadata & Validation                        │
│              • Alias Resolution                             │
│              • Context Integration                          │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                 Command Handlers                            │
│              • Individual cmd_*.go files                    │
│              • Context-aware operations                     │
│              • Error handling & validation                  │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                  Storage Interface                          │
│              • LMDB Operations                              │
│              • TTL Management                               │
│              • Context cancellation                        │
└─────────────────────┬───────────────────────────────────────┘
                      │
┌─────────────────────▼───────────────────────────────────────┐
│                    LMDB Backend                             │
│              • Memory-mapped storage                        │
│              • ACID transactions                            │
│              • Performance optimizations                    │
└─────────────────────────────────────────────────────────────┘
```

## Core Components

### 1. Command Registry (`internal/server/command_registry.go`)

The Command Registry is the central dispatch system that manages all Redis commands.

**Key Features:**
- **Metadata-driven validation**: Each command has defined min/max arguments
- **Alias support**: Commands can have multiple names (e.g., DELETE → DEL)
- **Context integration**: All handlers support cancellation
- **Extensible design**: Adding new commands requires minimal code changes

**Structure:**
```go
type CommandMetadata struct {
    Name     string
    MinArgs  int
    MaxArgs  int
    Handler  CommandHandler
    Aliases  []string
}
```

**Adding New Commands:**
1. Define metadata in `registerCommands()`
2. Create handler in `cmd_newcommand.go`
3. Add to appropriate handler map in `executeCommand()`

### 2. Command Handlers (`internal/server/cmd_*.go`)

Each Redis command is implemented in its own file following consistent patterns.

**Handler Signature:**
```go
func (server *Server) handleCommandName(ctx context.Context, conn redcon.Conn, cmd redcon.Command)
```

**Current Commands:**
- `PING` - Connection testing
- `SET` - Store key-value pairs
- `GET` - Retrieve values
- `DEL` - Delete keys
- `EXPIRE` - Set key expiration
- `EXPIREAT` - Set absolute expiration
- `TTL` - Get time to live
- `PTTL` - Get time to live in milliseconds
- `PERSIST` - Remove expiration

**Handler Pattern:**
1. Validate arguments using registry metadata
2. Extract command parameters
3. Call storage layer with context
4. Handle errors and cancellation
5. Write response to connection

### 3. Storage Layer (`internal/storage/`)

The storage layer provides a clean interface over LMDB with TTL management.

**Components:**
- `lmdb.go` - Core LMDB operations
- `ttl_manager.go` - TTL lifecycle management
- `ttl_storage.go` - TTL persistence
- `ttl_metrics.go` - Performance monitoring

**Key Features:**
- **Context support**: All operations can be canceled
- **TTL management**: Automatic expiration handling
- **Performance optimization**: Object pooling, batch operations
- **Metrics**: Comprehensive performance tracking

### 4. TTL Management

TTL (Time To Live) is implemented as a separate concern with its own storage.

**Architecture:**
```
Main Storage (keys → values)
     ↓
TTL Storage (keys → expiration_timestamp)
     ↓
Background Cleanup (periodic expiration)
```

**Components:**
- **TTLManager**: Interface for TTL operations
- **TTLStorage**: Persistence layer for TTL data
- **TTLMetrics**: Performance monitoring
- **Background cleanup**: Periodic expired key removal

## Design Patterns

### 1. Registry Pattern

The Command Registry implements a metadata-driven approach to command management:

```go
// Centralized command metadata
var commands = map[string]*CommandMetadata{
    "SET": {MinArgs: 3, MaxArgs: 3, Handler: handleSet},
}

// Automatic validation
err := registry.ValidateCommand(cmd, metadata)
```

**Benefits:**
- Eliminates code duplication
- Consistent validation
- Easy to add new commands
- Centralized configuration

### 2. Strategy Pattern

Command handlers implement the Strategy pattern through map dispatch:

```go
handlers := map[string]func(context.Context, redcon.Conn, redcon.Command){
    "SET": server.handleSet,
    "GET": server.handleGet,
}

handler := handlers[commandName]
handler(ctx, conn, cmd)
```

**Benefits:**
- No switch statements (follows code standards)
- Easy to extend
- Clean separation of concerns

### 3. Context Pattern

All operations support cancellation through Go's context package:

```go
func (server *Server) handleSet(ctx context.Context, conn redcon.Conn, cmd redcon.Command) {
    err := server.storage.SetWithContext(ctx, key, value)
    if isContextCanceled(err) {
        conn.WriteError("ERR operation canceled")
        return
    }
}
```

**Benefits:**
- Graceful cancellation
- Timeout support
- Resource cleanup
- Better user experience

## Performance Optimizations

### 1. LMDB Configuration

Aggressive performance flags for maximum throughput:

```go
PerformanceFlags = lmdb.WriteMap | lmdb.NoMetaSync | lmdb.NoSync | lmdb.MapAsync | lmdb.NoReadahead
```

### 2. Object Pooling

TTL metadata and byte slices are pooled to reduce GC pressure:

```go
var ttlMetadataPool = sync.Pool{
    New: func() interface{} {
        return &TTLMetadata{}
    },
}
```

### 3. Batch Operations

TTL cleanup processes keys in batches to avoid blocking:

```go
const MaxCleanupBatchSize = 1000
```

### 4. Atomic Metrics

Lock-free performance monitoring using atomic operations:

```go
atomic.AddInt64(&metrics.keysExpired, int64(count))
```

## Code Standards Compliance

The codebase follows strict Go standards for maintainability:

### 1. Naming Conventions
- **Descriptive receivers**: `func (server *Server)` not `func (s *Server)`
- **Clear function names**: `isContextCanceled()` not `isCanceled()`

### 2. Control Flow
- **Return early**: No `if/else` statements
- **Map dispatch**: No `switch` statements
- **Extracted conditions**: `isEmpty(key)` not `len(key) == 0`

### 3. File Organization
- **Single responsibility**: Each command in separate file
- **Utils separation**: Independent functions in `utils.go`
- **Clear structure**: Consistent file naming

### 4. Error Handling
- **Extracted checks**: `hasError(err)` not `err != nil`
- **Context awareness**: `isContextCanceled(err)`
- **Consistent patterns**: Same error handling across handlers

## Testing Strategy

### 1. Unit Tests
- Command registry validation
- Individual handler testing
- Storage layer operations
- TTL management

### 2. Integration Tests
- Full server with Redis client
- TTL persistence across restarts
- Performance benchmarks
- Error scenarios

### 3. Property-Based Tests
- TTL correctness properties
- Command argument validation
- Storage consistency

## Scalability Considerations

### 1. Command Addition
Adding new Redis commands is straightforward:

1. **Define metadata** in command registry
2. **Create handler** following existing patterns
3. **Add tests** for validation
4. **Update documentation**

### 2. Performance Scaling
- **LMDB memory mapping** scales with available RAM
- **Background cleanup** adapts to TTL density
- **Object pooling** reduces GC overhead
- **Atomic metrics** provide lock-free monitoring

### 3. Feature Extensions
- **Pub/Sub**: Can be added as new command category
- **Transactions**: LMDB supports ACID transactions
- **Clustering**: Architecture supports horizontal scaling
- **Replication**: Can be implemented at storage layer

## Monitoring and Observability

### 1. TTL Metrics
- Cleanup operations count
- Keys expired count
- Average cleanup duration
- Error rates

### 2. Performance Metrics
- Operations per second
- Latency percentiles
- Memory usage
- Connection counts

### 3. Health Checks
- LMDB status
- TTL manager health
- Background cleanup status

## Future Enhancements

### 1. Command Pattern Implementation
Consider implementing full Command Pattern for better extensibility:

```go
type Command interface {
    Execute(ctx context.Context, conn redcon.Conn, args [][]byte) error
    Validate(args [][]byte) error
}
```

### 2. Plugin System
Dynamic command loading for custom extensions:

```go
type CommandPlugin interface {
    Name() string
    Handler() CommandHandler
    Metadata() CommandMetadata
}
```

### 3. Advanced TTL Features
- **TTL callbacks**: Execute functions on expiration
- **Conditional TTL**: TTL based on access patterns
- **TTL inheritance**: Automatic TTL for related keys

## Conclusion

Keyp's architecture provides a solid foundation for a high-performance Redis-compatible server. The design emphasizes:

- **Maintainability**: Clear separation of concerns
- **Performance**: Optimized LMDB usage and object pooling
- **Extensibility**: Easy addition of new commands
- **Reliability**: Comprehensive testing and error handling
- **Standards compliance**: Consistent Go best practices

The architecture is ready for production use and can scale to support the full Redis command set while maintaining performance and code quality.