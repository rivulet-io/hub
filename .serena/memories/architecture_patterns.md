# Hub Project Architecture and Patterns

## Core Architecture

### Hub as Central Component
The `Hub` struct serves as the main entry point and orchestrator for all operations:
- Embeds a NATS server instance
- Manages JetStream, KV Store, and Object Store
- Provides unified interface for different messaging patterns
- Handles lifecycle management (creation, shutdown)

### Messaging Patterns Hierarchy

#### 1. Volatile Messaging (Real-time, No Persistence)
- **Fanout**: `SubscribeVolatileViaFanout` - All subscribers receive messages
- **Queue**: `SubscribeVolatileViaQueue` - Load balancing among subscribers
- **Request/Reply**: `RequestVolatile` - Synchronous request-response

#### 2. Persistent Messaging (JetStream)
- **Durable Consumers**: `SubscribePersistentViaDurable` - Survive restarts
- **Ephemeral Consumers**: `SubscribePersistentViaEphemeral` - Temporary
- **Streams**: `CreateOrUpdatePersistent` - Configure message storage

#### 3. Storage Patterns
- **Key-Value Store**: For configuration, settings, caching
- **Object Store**: For large files, documents, media

## Configuration Strategy

### Options Pattern
- Single `Options` struct contains all configuration
- Separate default constructors for different deployment types:
  - `DefaultNodeOptions()` - Single node setup
  - `DefaultGatewayOptions()` - Gateway configuration  
  - `DefaultLeafOptions()` - Leaf node configuration

### Size Management
- Custom `Size` type with human-readable constructors
- Supports bytes to exabytes with proper formatting
- Used throughout for memory, storage, and payload limits

## Error Handling Strategy
- Consistent error return as last parameter
- Descriptive error messages with context
- Error wrapping using `fmt.Errorf` with `%w`
- Separate error handlers for async operations

## Handler Function Design
Different handler signatures for different use cases:
- Volatile handlers: `(subject, msg) -> (response, shouldReply)`
- Persistent handlers: `(subject, msg) -> (response, shouldReply, shouldAck)`
- Error handlers: `(error) -> void`

## Testing Architecture
- Unit tests for each component (`*_test.go`)
- Integration tests in separate `tests/` directory
- Comprehensive scenario coverage including error cases
- Performance and concurrency testing included

## Design Patterns Used

### Factory Pattern
- `NewHub()` for Hub creation
- `New*Options()` for configuration creation
- `NewSizeFrom*()` for size construction

### Builder Pattern
- Configuration structs build up complex options
- Separate configs for different storage types

### Observer Pattern
- Handler functions act as observers for message events
- Error handlers for error event observation

### Strategy Pattern
- Different retention policies for streams
- Different authentication methods
- Multiple messaging patterns for different use cases