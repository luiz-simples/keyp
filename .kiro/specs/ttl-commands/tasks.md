# Implementation Plan: TTL Commands

## Overview

Implementation of TTL (Time To Live) system for Keyp server with Redis-compatible commands: EXPIRE, EXPIREAT, TTL, PTTL, and PERSIST. The implementation follows a modular approach with separate TTL manager, storage integration, and command handlers.

## Tasks

- [x] 1. Set up TTL storage infrastructure
  - Create TTL metadata storage using LMDB
  - Implement TTLStorage interface with atomic operations
  - Add TTL-specific error types and constants
  - _Requirements: 5.1, 5.2_

- [x] 1.1 Write property test for TTL storage operations
  - **Property 5: TTL Persistence Round-trip**
  - **Validates: Requirements 5.1, 5.2**

- [x] 1.2 Write integration tests for TTL storage with go-redis
  - Test TTL metadata persistence via Redis protocol
  - Validate storage consistency across server restarts
  - Test concurrent TTL operations via multiple Redis clients
  - _Requirements: 5.1, 5.2_

- [ ] 2. Implement TTL Manager core functionality
  - [x] 2.1 Create TTLManager struct and interface
    - Implement SetExpire and SetExpireAt methods
    - Add time utilities and validation functions
    - _Requirements: 1.1, 1.2_

  - [x] 2.2 Write property test for TTL setting consistency
    - **Property 1: TTL Setting Consistency**
    - **Validates: Requirements 1.1, 1.2**

  - [x] 2.3 Write integration tests for EXPIRE/EXPIREAT with go-redis
    - Test EXPIRE command via Redis client with various TTL values
    - Test EXPIREAT command via Redis client with Unix timestamps
    - Validate Redis protocol compatibility and response codes
    - Test edge cases: negative TTL, non-existent keys, large TTL values
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.4 Implement TTL query methods (GetTTL, GetPTTL)
    - Add remaining time calculation logic
    - Handle persistent and non-existent key cases
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 2.5 Write property test for TTL query accuracy
    - **Property 2: TTL Query Accuracy**
    - **Validates: Requirements 2.1, 2.4**

  - [x] 2.6 Write integration tests for TTL/PTTL with go-redis
    - Test TTL command via Redis client for keys with expiration
    - Test PTTL command via Redis client with millisecond precision
    - Validate return codes: -1 for persistent, -2 for non-existent
    - Test TTL accuracy over time with multiple queries
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 2.7 Implement Persist method
    - Add TTL removal functionality
    - Return appropriate status codes
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 2.8 Write property test for persist operation
    - **Property 3: Persist Operation Idempotency**
    - **Validates: Requirements 3.1**

  - [x] 2.9 Write integration tests for PERSIST with go-redis
    - Test PERSIST command via Redis client on keys with TTL
    - Test PERSIST command on persistent and non-existent keys
    - Validate return codes and TTL removal behavior
    - Test PERSIST idempotency via multiple calls
    - _Requirements: 3.1, 3.2, 3.3_

- [ ] 3. Implement automatic expiration system
  - [ ] 3.1 Create expiration index for efficient cleanup
    - Implement time-based indexing structure
    - Add concurrent access protection
    - _Requirements: 4.1_

  - [ ] 3.2 Implement lazy expiration on key access
    - Add IsExpired method to TTL manager
    - Integrate expiration checks in storage operations
    - _Requirements: 4.2, 4.3_

  - [ ]* 3.3 Write property test for expiration consistency
    - **Property 4: Expiration Consistency**
    - **Validates: Requirements 4.1, 4.2, 4.3**

  - [ ]* 3.4 Write integration tests for automatic expiration with go-redis
    - Test key expiration via Redis client with short TTLs
    - Test lazy expiration on GET operations via Redis protocol
    - Test that expired keys return null via Redis client
    - Test cleanup behavior with multiple expired keys
    - _Requirements: 4.1, 4.2, 4.3_

  - [ ] 3.5 Implement background cleanup process
    - Create cleanup goroutine with configurable intervals
    - Add batch cleanup for expired keys
    - _Requirements: 4.1, 4.4_

- [ ] 4. Checkpoint - Core TTL functionality complete
  - Ensure all TTL manager tests pass, ask the user if questions arise.

- [ ] 5. Integrate TTL system with existing storage
  - [ ] 5.1 Enhance LMDBStorage with TTL support
    - Add TTL metadata database to LMDB setup
    - Implement TTL-aware Get/Set/Del operations
    - _Requirements: 5.1, 5.2_

  - [ ] 5.2 Update existing commands to handle TTL
    - Modify GET command to check expiration
    - Update SET command to preserve existing TTL
    - Update DEL command to clean TTL metadata
    - _Requirements: 4.2, 4.3_

  - [ ]* 5.3 Write integration tests for TTL storage
    - Test TTL persistence across restarts
    - Test cleanup during startup
    - _Requirements: 5.3_

  - [ ]* 5.4 Write comprehensive integration tests with go-redis
    - Test TTL integration with existing SET/GET/DEL commands
    - Test TTL behavior with Redis client during server restart
    - Test concurrent TTL operations from multiple Redis clients
    - Validate TTL metadata consistency via Redis protocol
    - _Requirements: 4.2, 4.3, 5.1, 5.2, 5.3_

- [ ] 6. Implement TTL command handlers
  - [x] 6.1 Create cmd_expire.go
    - Implement handleExpire with argument validation
    - Add seconds parsing and validation
    - _Requirements: 1.1, 6.1, 6.2, 6.3_

  - [ ] 6.2 Create cmd_expireat.go
    - Implement handleExpireAt with timestamp validation
    - Add Unix timestamp parsing and validation
    - _Requirements: 1.2, 6.1, 6.4_

  - [ ] 6.3 Create cmd_ttl.go
    - Implement handleTTL with proper return codes
    - Handle persistent and non-existent key cases
    - _Requirements: 2.1, 2.2, 2.3, 6.1_

  - [ ] 6.4 Create cmd_pttl.go
    - Implement handlePTTL with millisecond precision
    - Handle persistent and non-existent key cases
    - _Requirements: 2.4, 2.5, 6.1_

  - [ ] 6.5 Create cmd_persist.go
    - Implement handlePersist with proper return codes
    - Handle already persistent and non-existent keys
    - _Requirements: 3.1, 3.2, 3.3, 6.1_

  - [ ]* 6.6 Write property test for command validation
    - **Property 6: Command Validation Consistency**
    - **Validates: Requirements 6.1**

  - [ ]* 6.7 Write integration tests for all TTL commands with go-redis
    - Test complete TTL command suite via Redis client
    - Test Redis protocol compatibility for all TTL commands
    - Test error messages and response formats match Redis
    - Test command argument validation via Redis protocol
    - Test TTL commands with various data types and edge cases
    - _Requirements: 1.1, 1.2, 2.1, 2.4, 3.1, 6.1, 6.2, 6.3, 6.4_

- [ ] 7. Add TTL command constants and utilities
  - [ ] 7.1 Update server/utils.go with TTL validation functions
    - Add TTL argument count validation functions
    - Add time parsing and validation utilities
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [ ] 7.2 Update server command dispatcher
    - Add TTL commands to handler map
    - Ensure proper command routing
    - _Requirements: All TTL commands_

- [ ] 8. Implement TTL persistence and recovery
  - [ ] 8.1 Add TTL restoration on server startup
    - Implement RestoreTTL method
    - Add startup cleanup for expired keys
    - _Requirements: 5.2, 5.3_

  - [ ] 8.2 Add graceful shutdown for TTL system
    - Ensure TTL metadata is properly persisted
    - Stop background cleanup processes cleanly
    - _Requirements: 5.1_

- [ ]* 8.3 Write property test for TTL persistence
  - Test TTL survival across server restarts
  - Test startup cleanup behavior
  - _Requirements: 5.2, 5.3_

- [ ]* 8.4 Write integration tests for TTL persistence with go-redis
  - Test TTL commands work correctly after server restart via Redis client
  - Test expired key cleanup during startup via Redis protocol
  - Test TTL restoration accuracy after restart via Redis client
  - Validate no data loss during restart with active TTLs
  - _Requirements: 5.1, 5.2, 5.3_

- [ ] 9. Performance optimization and monitoring
  - [ ] 9.1 Add TTL performance metrics
    - Track cleanup performance and frequency
    - Monitor TTL storage overhead
    - _Requirements: 4.1, 4.4_

  - [ ] 9.2 Optimize expiration index performance
    - Implement efficient time-based lookups
    - Add batch operations for cleanup
    - _Requirements: 4.1_

- [ ]* 9.3 Write benchmark tests for TTL operations
  - Benchmark TTL set/get/cleanup operations
  - Test performance with large numbers of keys
  - _Requirements: Performance_

- [ ]* 9.4 Write performance integration tests with go-redis
  - Benchmark TTL commands via Redis client under load
  - Test TTL system performance with concurrent Redis clients
  - Measure TTL operation latency via Redis protocol
  - Test cleanup performance with large datasets via Redis client
  - _Requirements: Performance, 4.1_

- [ ] 10. Final checkpoint - Complete TTL system
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- **Integration tests with go-redis** validate Redis protocol compatibility and real-world usage
- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- Unit tests validate specific examples and edge cases
- **go-redis integration tests** ensure end-to-end functionality via Redis client
- TTL system integrates seamlessly with existing Keyp architecture
- **Redis compatibility** is validated through comprehensive go-redis test suite