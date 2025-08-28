# Hub

Hub is a powerful data communicator library for Go applications that embeds a NATS server and provides high-level abstractions for JetStream, Key-Value Store, and Object Store operations.

## Features

- **Embedded NATS Server**: Run a full NATS server in-process without external dependencies
- **JetStream Support**: Persistent messaging with streams, durable/ephemeral subscriptions
- **Key-Value Store**: Distributed key-value storage with versioning and TTL support
- **Object Store**: Large object storage with metadata support
- **Volatile Messaging**: Standard NATS publish/subscribe and request-reply patterns
- **Clustering**: Built-in support for NATS clustering
- **Size Utilities**: Convenient size handling with human-readable formats

## Installation

```bash
go get github.com/snowmerak/hub
```

## Quick Start

```go
package main

import (
    "fmt"
    "time"

    "github.com/snowmerak/hub"
)

func main() {
    // Create default options
    opts, err := hub.DefaultOptions()
    if err != nil {
        panic(err)
    }

    // Create and start hub
    h, err := hub.NewHub(opts)
    if err != nil {
        panic(err)
    }
    defer h.Shutdown()

    // Volatile messaging example
    cancel, err := h.SubscribeVolatileViaFanout("greetings", func(subject string, msg []byte) ([]byte, bool) {
        fmt.Printf("Received: %s\n", string(msg))
        return []byte("Hello back!"), true
    }, func(err error) {
        fmt.Printf("Error: %v\n", err)
    })
    if err != nil {
        panic(err)
    }
    defer cancel()

    // Publish a message
    err = h.PublishVolatile("greetings", []byte("Hello, Hub!"))
    if err != nil {
        panic(err)
    }

    time.Sleep(time.Second)
}
```

## Usage Examples

### JetStream Persistent Messaging

```go
// Create a persistent stream
config := &hub.PersistentConfig{
    Subjects:  []string{"orders.>"},
    Retention: nats.LimitsPolicy,
    MaxMsgs:   1000000,
}
err := h.CreateOrUpdatePersistent(config)
if err != nil {
    panic(err)
}

// Subscribe with durable consumer
cancel, err := h.SubscribePersistentViaDurable("order-processor", "orders.new", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("Processing order: %s\n", string(msg))
    return nil, false, true // no reply, ack message
}, func(err error) {
    fmt.Printf("Error: %v\n", err)
})
```

### Key-Value Store

```go
// Create KV store
kvConfig := hub.KeyValueStoreConfig{
    Bucket:       "user-data",
    Description:  "User preferences and settings",
    MaxValueSize: hub.NewSizeFromMegabytes(1),
    TTL:          24 * time.Hour,
}
err := h.CreateOrUpdateKeyValueStore(kvConfig)
if err != nil {
    panic(err)
}

// Put and get values
revision, err := h.PutToKeyValueStore("user-data", "user123", []byte("preferences data"))
if err != nil {
    panic(err)
}

data, rev, err := h.GetFromKeyValueStore("user-data", "user123")
if err != nil {
    panic(err)
}
fmt.Printf("Data: %s, Revision: %d\n", string(data), rev)
```

### Object Store

```go
// Create object store
objConfig := hub.ObjectStoreConfig{
    Bucket:      "documents",
    Description: "User uploaded documents",
    MaxBytes:    hub.NewSizeFromGigabytes(10),
}
err := h.CreateObjectStore(objConfig)
if err != nil {
    panic(err)
}

// Store and retrieve objects
err = h.PutToObjectStore("documents", "report.pdf", pdfData, map[string]string{
    "filename": "monthly-report.pdf",
    "type":     "application/pdf",
})
if err != nil {
    panic(err)
}

data, err := h.GetFromObjectStore("documents", "report.pdf")
if err != nil {
    panic(err)
}
```

## Configuration

The `Options` struct provides comprehensive configuration:

```go
opts := &hub.Options{
    Name:               "my-hub",
    Host:               "0.0.0.0",
    Port:               4222,
    AuthorizationToken: "your-token",
    MaxPayload:         hub.NewSizeFromMegabytes(8),

    // Clustering options
    ClusterHost:         "0.0.0.0",
    ClusterPort:         6222,
    ClusterUsername:     "cluster-user",
    ClusterPassword:     "cluster-pass",

    // JetStream options
    JetstreamMaxMemory:  hub.NewSizeFromMegabytes(512),
    JetstreamMaxStorage: hub.NewSizeFromGigabytes(10),
    StoreDir:            "./data",

    // Logging
    LogFile:      "./nats.log",
    LogSizeLimit: 10 * 1024 * 1024,
}
```

## API Reference

### Core Methods
- `NewHub(opts *Options) (*Hub, error)` - Create and start a new hub
- `Shutdown()` - Gracefully shutdown the hub

### Volatile Messaging
- `SubscribeVolatileViaFanout(subject, handler, errHandler)` - Subscribe with fanout
- `SubscribeVolatileViaQueue(subject, queue, handler, errHandler)` - Subscribe with queue group
- `PublishVolatile(subject, msg)` - Publish a message
- `RequestVolatile(subject, msg, timeout)` - Send request and wait for response

### JetStream
- `CreateOrUpdatePersistent(config)` - Create/update persistent stream
- `SubscribePersistentViaDurable(id, subject, handler, errHandler)` - Durable subscription
- `SubscribePersistentViaEphemeral(subject, handler, errHandler)` - Ephemeral subscription
- `PublishPersistent(subject, msg)` - Publish to stream

### Key-Value Store
- `CreateOrUpdateKeyValueStore(config)` - Create KV store
- `GetFromKeyValueStore(bucket, key)` - Get value
- `PutToKeyValueStore(bucket, key, value)` - Put value
- `UpdateToKeyValueStore(bucket, key, value, expectedRevision)` - Update with revision check
- `DeleteFromKeyValueStore(bucket, key)` - Delete key

### Object Store
- `CreateObjectStore(config)` - Create object store
- `GetFromObjectStore(bucket, key)` - Get object
- `PutToObjectStore(bucket, key, data, metadata)` - Store object
- `DeleteFromObjectStore(bucket, key)` - Delete object

## Size Utilities

```go
// Create sizes
size := hub.NewSizeFromMegabytes(512)
size = hub.NewSizeFromGigabytes(10)

// Convert and format
fmt.Println(size.String()) // "10.00 GB"
bytes := size.Bytes()      // int64 value
```

## Requirements

- Go 1.25.0 or later
- Linux, macOS, or Windows

## Dependencies

- [NATS Server](https://github.com/nats-io/nats-server) - Embedded NATS server
- [NATS Go Client](https://github.com/nats-io/nats.go) - NATS client library
- [Randflake](https://gosuda.org/randflake) - ID generation

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
