# Requirements Document

## Introduction

Sistema de TTL (Time To Live) para o servidor Keyp que permite definir tempo de expiração para chaves, verificar tempo restante e gerenciar expiração automática de dados.

## Glossary

- **TTL_System**: Sistema responsável por gerenciar tempo de vida das chaves
- **Expiration_Time**: Timestamp Unix quando uma chave deve expirar
- **TTL_Value**: Tempo em segundos até a expiração de uma chave
- **Expired_Key**: Chave que ultrapassou seu tempo de expiração
- **Persistent_Key**: Chave sem tempo de expiração definido

## Requirements

### Requirement 1: Set Key Expiration

**User Story:** As a developer, I want to set expiration time for keys, so that data can be automatically cleaned up after a specified period.

#### Acceptance Criteria

1. WHEN a user executes EXPIRE command with valid key and seconds, THE TTL_System SHALL set the expiration time for the key
2. WHEN a user executes EXPIREAT command with valid key and Unix timestamp, THE TTL_System SHALL set the absolute expiration time for the key
3. WHEN a user tries to set expiration on non-existent key, THE TTL_System SHALL return zero indicating failure
4. WHEN a user sets expiration with negative seconds, THE TTL_System SHALL immediately expire the key
5. WHEN a user sets expiration on already expired key, THE TTL_System SHALL return zero indicating failure

### Requirement 2: Query Key TTL

**User Story:** As a developer, I want to check remaining time for keys, so that I can monitor expiration status.

#### Acceptance Criteria

1. WHEN a user executes TTL command on key with expiration, THE TTL_System SHALL return remaining seconds until expiration
2. WHEN a user executes TTL command on persistent key, THE TTL_System SHALL return -1 indicating no expiration
3. WHEN a user executes TTL command on non-existent key, THE TTL_System SHALL return -2 indicating key not found
4. WHEN a user executes PTTL command on key with expiration, THE TTL_System SHALL return remaining milliseconds until expiration
5. WHEN a user executes PTTL command on persistent key, THE TTL_System SHALL return -1 indicating no expiration

### Requirement 3: Remove Key Expiration

**User Story:** As a developer, I want to remove expiration from keys, so that they become persistent again.

#### Acceptance Criteria

1. WHEN a user executes PERSIST command on key with expiration, THE TTL_System SHALL remove the expiration and return 1
2. WHEN a user executes PERSIST command on persistent key, THE TTL_System SHALL return 0 indicating no change
3. WHEN a user executes PERSIST command on non-existent key, THE TTL_System SHALL return 0 indicating failure

### Requirement 4: Automatic Key Expiration

**User Story:** As a system administrator, I want keys to be automatically removed when expired, so that memory usage is optimized.

#### Acceptance Criteria

1. WHEN a key reaches its expiration time, THE TTL_System SHALL automatically remove the key from storage
2. WHEN an expired key is accessed via GET command, THE TTL_System SHALL return null and remove the key
3. WHEN an expired key is accessed via any command, THE TTL_System SHALL treat it as non-existent
4. WHEN TTL_System performs cleanup, THE TTL_System SHALL maintain data integrity during removal process

### Requirement 5: TTL Persistence

**User Story:** As a system administrator, I want TTL information to survive server restarts, so that expiration behavior is consistent.

#### Acceptance Criteria

1. WHEN server stores a key with TTL, THE TTL_System SHALL persist both key data and expiration metadata
2. WHEN server restarts, THE TTL_System SHALL restore all TTL information from persistent storage
3. WHEN server starts up, THE TTL_System SHALL immediately clean up any keys that expired during downtime
4. WHEN TTL metadata is corrupted, THE TTL_System SHALL treat affected keys as persistent and log the issue

### Requirement 6: TTL Command Validation

**User Story:** As a developer, I want proper validation of TTL commands, so that I receive clear error messages for invalid usage.

#### Acceptance Criteria

1. WHEN a user provides invalid number of arguments to TTL commands, THE TTL_System SHALL return appropriate error message
2. WHEN a user provides non-numeric TTL values, THE TTL_System SHALL return error indicating invalid argument type
3. WHEN a user provides TTL values outside valid range, THE TTL_System SHALL return error with acceptable range information
4. WHEN a user provides malformed Unix timestamp, THE TTL_System SHALL return error indicating invalid timestamp format