# Hub Project Overview

## Purpose
Hub is a powerful data communicator library for Go applications that embeds a NATS server and provides high-level abstractions for messaging patterns including:

- **Embedded NATS Server**: Run a full NATS server in-process without external dependencies
- **JetStream Support**: Persistent messaging with streams, durable/ephemeral subscriptions
- **Key-Value Store**: Distributed key-value storage with versioning and TTL support
- **Object Store**: Large object storage with metadata support
- **Volatile Messaging**: Standard NATS publish/subscribe and request-reply patterns
- **Clustering**: Built-in support for NATS clustering
- **Size Utilities**: Convenient size handling with human-readable formats

## Tech Stack
- **Language**: Go 1.25.0+
- **Dependencies**:
  - `github.com/nats-io/nats-server/v2 v2.11.8` - Embedded NATS server
  - `github.com/nats-io/nats.go v1.45.0` - NATS client library
- **Module**: `github.com/snowmerak/hub`

## Project Structure
The codebase is organized into focused files:

- `hub.go` - Core Hub struct and creation logic, main Options configuration
- `core.go` - Volatile messaging (basic pub/sub, queue, request/reply)
- `jetstream.go` - JetStream operations (persistent messaging)
- `kv_store.go` - Key-Value store operations
- `object_store.go` - Object store operations
- `auth.go` - Authentication methods and configurations
- `size.go` - Size utilities with human-readable formats
- `*_test.go` - Comprehensive test coverage for all components
- `tests/` - Integration tests directory with extensive test scenarios

## Key Features
1. **Multiple Messaging Patterns**: Fanout, Queue-based, Request/Reply, Persistent JetStream
2. **Data Storage**: Key-Value and Object stores with metadata support
3. **Clustering Support**: Built-in NATS clustering capabilities
4. **Size Management**: Convenient size handling (bytes to exabytes)
5. **Authentication**: Custom authentication methods
6. **Persistence**: JetStream with configurable retention policies