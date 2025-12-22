# Design Document: TTL Commands

## Overview

This design implements a comprehensive TTL (Time To Live) system for the Keyp server, adding Redis-compatible expiration functionality. The system supports setting expiration times, querying remaining TTL, removing expiration, and automatic cleanup of expired keys.

## Architecture

### Core Components

```
TTL System Architecture:
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   TTL Commands  │    │  TTL Manager    │    │  TTL Storage    │
│                 │    │                 │    │                 │
│ • EXPIRE        │───▶│ • Set TTL       │───▶│ • Metadata DB   │
│ • EXPIREAT      │    │ • Query TTL     │    │ • Expiry Index  │
│ • TTL/PTTL      │    │ • Remove TTL    │    │ • Cleanup Queue │
│ • PERSIST       │    │ • Auto Cleanup  │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### TTL Manager Responsibilities
- **Expiration Management**: Set, query, and remove TTL for keys
- **Automatic Cleanup**: Background process to remove expired keys
- **Lazy Expiration**: Check and remove expired keys on access
- **Persistence**: Store and restore TTL metadata across restarts

### Storage Layer Integration
- **Dual Storage**: Main key-value storage + TTL metadata storage
- **Atomic Operations**: Ensure consistency between data and TTL metadata
- **Efficient Indexing**: Time-based index for efficient expiration cleanup

## Components and Interfaces

### TTL Manager Interface

```go
type TTLManager interface {
    // Set expiration in seconds from now
    SetExpire(key []byte, seconds int64) (int, error)
    
    // Set expiration at absolute Unix timestamp
    SetExpireAt(key []byte, timestamp int64) (int, error)
    
    // Get remaining TTL in seconds (-1 = persistent, -2 = not found)
    GetTTL(key []byte) (int64, error)
    
    // Get remaining TTL in milliseconds (-1 = persistent, -2 = not found)
    GetPTTL(key []byte) (int64, error)
    
    // Remove expiration, make key persistent
    Persist(key []byte) (int, error)
    
    // Check if key is expired (for lazy expiration)
    IsExpired(key []byte) (bool, error)
    
    // Clean up expired keys (background process)
    CleanupExpired() error
    
    // Restore TTL data after restart
    RestoreTTL() error
}
```

### TTL Storage Schema

```go
type TTLMetadata struct {
    Key        []byte    // The key this TTL applies to
    ExpiresAt  int64     // Unix timestamp when key expires
    CreatedAt  int64     // When TTL was set (for debugging)
}

type TTLStorage interface {
    // Store TTL metadata
    SetTTL(key []byte, expiresAt int64) error
    
    // Get TTL metadata
    GetTTL(key []byte) (*TTLMetadata, error)
    
    // Remove TTL metadata
    RemoveTTL(key []byte) error
    
    // Get all keys expiring before timestamp
    GetExpiredKeys(before int64) ([][]byte, error)
    
    // Batch remove TTL metadata
    RemoveTTLBatch(keys [][]byte) error
}
```

### Command Handlers

```go
// cmd_expire.go
func (server *Server) handleExpire(conn redcon.Conn, cmd redcon.Command) {
    // Validate arguments: EXPIRE key seconds
    // Parse seconds as integer
    // Call TTLManager.SetExpire()
    // Return 1 if successful, 0 if key doesn't exist
}

// cmd_expireat.go  
func (server *Server) handleExpireAt(conn redcon.Conn, cmd redcon.Command) {
    // Validate arguments: EXPIREAT key timestamp
    // Parse timestamp as Unix timestamp
    // Call TTLManager.SetExpireAt()
    // Return 1 if successful, 0 if key doesn't exist
}

// cmd_ttl.go
func (server *Server) handleTTL(conn redcon.Conn, cmd redcon.Command) {
    // Validate arguments: TTL key
    // Call TTLManager.GetTTL()
    // Return seconds, -1 for persistent, -2 for not found
}

// cmd_pttl.go
func (server *Server) handlePTTL(conn redcon.Conn, cmd redcon.Command) {
    // Validate arguments: PTTL key
    // Call TTLManager.GetPTTL()
    // Return milliseconds, -1 for persistent, -2 for not found
}

// cmd_persist.go
func (server *Server) handlePersist(conn redcon.Conn, cmd redcon.Command) {
    // Validate arguments: PERSIST key
    // Call TTLManager.Persist()
    // Return 1 if TTL removed, 0 if key was already persistent or doesn't exist
}
```

## Data Models

### TTL Metadata Storage

```go
// TTL metadata stored separately from main key-value data
type TTLEntry struct {
    Key       []byte  // Original key
    ExpiresAt int64   // Unix timestamp (seconds)
}

// In-memory expiration index for efficient cleanup
type ExpirationIndex struct {
    // Time-ordered map: timestamp -> []keys
    timeIndex map[int64][][]byte
    
    // Key lookup: key -> timestamp  
    keyIndex  map[string]int64
    
    // Mutex for concurrent access
    mutex     sync.RWMutex
}
```

### Storage Integration

```go
// Enhanced storage interface with TTL support
type StorageWithTTL interface {
    // Existing methods
    Set(key, value []byte) error
    Get(key []byte) ([]byte, error)
    Del(keys ...[]byte) (int, error)
    
    // TTL-aware methods
    GetWithTTLCheck(key []byte) ([]byte, bool, error) // value, expired, error
    SetWithTTL(key, value []byte, ttl *TTLMetadata) error
    
    // TTL metadata methods
    SetTTLMetadata(key []byte, expiresAt int64) error
    GetTTLMetadata(key []byte) (*TTLMetadata, error)
    RemoveTTLMetadata(key []byte) error
    GetExpiredKeys(before int64) ([][]byte, error)
}
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property Reflection

After reviewing the prework analysis, I identified several properties that can be consolidated:
- Properties 2.1 and 2.4 (TTL and PTTL queries) can be combined into one comprehensive time query property
- Properties 4.1, 4.2, and 4.3 (automatic expiration behaviors) can be combined into one expiration consistency property
- Properties 5.1 and 5.2 (persistence behaviors) can be combined into one TTL persistence property

### Core TTL Properties

**Property 1: TTL Setting Consistency**
*For any* existing key and positive TTL value, setting expiration should result in the key having the specified TTL within reasonable time bounds
**Validates: Requirements 1.1, 1.2**

**Property 2: TTL Query Accuracy**  
*For any* key with TTL set, querying TTL should return a value that decreases over time and matches the originally set expiration within reasonable bounds
**Validates: Requirements 2.1, 2.4**

**Property 3: Persist Operation Idempotency**
*For any* key with TTL, applying PERSIST should remove the TTL, and applying PERSIST again should have no additional effect
**Validates: Requirements 3.1**

**Property 4: Expiration Consistency**
*For any* key that has reached its expiration time, the key should be treated as non-existent across all operations and should be automatically removed from storage
**Validates: Requirements 4.1, 4.2, 4.3**

**Property 5: TTL Persistence Round-trip**
*For any* key with TTL, storing the key with TTL metadata and then restarting the system should preserve both the key data and TTL information accurately
**Validates: Requirements 5.1, 5.2, 5.3**

**Property 6: Command Validation Consistency**
*For any* TTL command with invalid arguments, the system should return appropriate error messages and not modify any key state
**Validates: Requirements 6.1**

## Error Handling

### TTL-Specific Errors

```go
var (
    ErrInvalidTTL       = errors.New("invalid TTL value")
    ErrInvalidTimestamp = errors.New("invalid timestamp")
    ErrTTLNotFound      = errors.New("TTL not found for key")
    ErrTTLCorrupted     = errors.New("TTL metadata corrupted")
)
```

### Error Response Patterns

- **Invalid Arguments**: Return error message with correct usage
- **Non-existent Keys**: Return 0 for SET operations, -2 for GET operations
- **Invalid TTL Values**: Return error for negative values (except EXPIRE with negative = immediate expiry)
- **Corrupted Metadata**: Log error, treat key as persistent, continue operation

### Graceful Degradation

- **TTL Storage Failure**: Fall back to treating all keys as persistent
- **Cleanup Process Failure**: Log error, continue serving requests
- **Metadata Corruption**: Isolate corrupted entries, preserve valid data

## Testing Strategy

### Dual Testing Approach

**Unit Tests**: Verify specific examples, edge cases, and error conditions
- Test specific TTL values (0, 1, 3600, negative values)
- Test edge cases (non-existent keys, already expired keys)
- Test error conditions (invalid arguments, corrupted data)
- Test integration between TTL manager and storage layer

**Property Tests**: Verify universal properties across all inputs
- Minimum 100 iterations per property test
- Generate random keys, TTL values, and timestamps
- Test TTL consistency across different time ranges
- Test persistence behavior with random restart scenarios
- Test cleanup efficiency with various expiration patterns

**Integration Tests with go-redis**: Validate Redis protocol compatibility and real-world usage
- Test all TTL commands via Redis client (github.com/redis/go-redis/v9)
- Validate response formats and error messages match Redis specification
- Test concurrent operations from multiple Redis clients
- Test TTL behavior during server restarts via Redis protocol
- Benchmark performance under realistic client load
- Ensure seamless compatibility with existing Redis tooling

### Property Test Configuration

Each property test must reference its design document property:
- **Feature: ttl-commands, Property 1**: TTL Setting Consistency
- **Feature: ttl-commands, Property 2**: TTL Query Accuracy  
- **Feature: ttl-commands, Property 3**: Persist Operation Idempotency
- **Feature: ttl-commands, Property 4**: Expiration Consistency
- **Feature: ttl-commands, Property 5**: TTL Persistence Round-trip
- **Feature: ttl-commands, Property 6**: Command Validation Consistency

### Integration Test Configuration

Each integration test must validate Redis compatibility:
- Use go-redis client for all protocol-level testing
- Test command syntax and response format compatibility
- Validate error codes and messages match Redis behavior
- Test concurrent access patterns typical in Redis usage
- Measure performance characteristics comparable to Redis

### Testing Challenges

- **Time-based Testing**: Use controllable time sources for deterministic tests
- **Concurrency Testing**: Test TTL operations under concurrent access via multiple Redis clients
- **Performance Testing**: Verify cleanup efficiency with large numbers of keys via go-redis benchmarks
- **Persistence Testing**: Test behavior across simulated restarts using Redis client connections
- **Protocol Compatibility**: Ensure exact Redis protocol compliance through comprehensive go-redis test suite