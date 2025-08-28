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

### 1. Simple Publish/Subscribe (Basic Messaging)

The most basic messaging pattern where a publisher sends messages and subscribers receive them.

```go
// Subscriber
cancel, err := h.SubscribeVolatileViaFanout("news.updates", func(subject string, msg []byte) ([]byte, bool) {
    fmt.Printf("Received news: %s\n", string(msg))
    return nil, false // no response
}, func(err error) {
    log.Printf("Subscription error: %v", err)
})

// Publisher
err = h.PublishVolatile("news.updates", []byte("New news update!"))
```

**Use cases**: Real-time notifications, log collection, event broadcasting

### 2. QueueSub (Queue-based Subscription)

A pattern where only one subscriber among multiple subscribers processes each message.

```go
// Multiple workers join the same queue group
for i := 0; i < 3; i++ {
    workerID := fmt.Sprintf("worker-%d", i)
    cancel, err := h.SubscribeVolatileViaQueue("tasks", "task-queue", func(subject string, msg []byte) ([]byte, bool) {
        fmt.Printf("Worker %s processing task: %s\n", workerID, string(msg))
        return nil, false
    }, func(err error) {
        log.Printf("Worker %s error: %v", workerID, err)
    })
    defer cancel()
}

// Publish task - only one worker will process it
err = h.PublishVolatile("tasks", []byte("Data processing task"))
```

**Use cases**: Task distribution, load balancing, microservice task processing

### 3. Request/Reply (Request-Response)

A synchronous request-response pattern.

```go
// Responder
cancel, err := h.SubscribeVolatileViaFanout("calculator.add", func(subject string, msg []byte) ([]byte, bool) {
    // Process request like "2+3" and calculate result
    result := calculate(string(msg))
    return []byte(fmt.Sprintf("Result: %d", result)), true // return response
}, func(err error) {
    log.Printf("Calculator error: %v", err)
})

// Requester
response, err := h.RequestVolatile("calculator.add", []byte("2+3"), 5*time.Second)
if err != nil {
    log.Printf("Request failed: %v", err)
} else {
    fmt.Printf("Calculation result: %s\n", string(response))
}
```

**Use cases**: RPC calls, inter-service communication, data queries

### 4. JetStream Persistent Messaging

Messages are stored on disk and persist even after system restarts.

```go
// Create stream
config := &hub.PersistentConfig{
    Description: "Order processing stream",
    Subjects:    []string{"orders.>"},
    Retention:   0, // Limits policy
    MaxMsgs:     100000,
    MaxAge:      24 * time.Hour,
}
err := h.CreateOrUpdatePersistent(config)

// Create durable subscription
cancel, err := h.SubscribePersistentViaDurable("order-processor", "orders.new", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("Processing order: %s\n", string(msg))
    return nil, false, true // send ACK
}, func(err error) {
    log.Printf("Order processing error: %v", err)
})

// Publish message
err = h.PublishPersistent("orders.new", []byte("New order: Product A x 2"))
```

**Use cases**: Order systems, event sourcing, audit logs

### 5. JetStream QueueSub (Persistent Queue)

Combines JetStream's persistent storage with queue-based processing.

```go
// Create stream (Limits policy allows multiple consumers)
config := &hub.PersistentConfig{
    Description: "Task queue stream",
    Subjects:    []string{"work.>"},
    Retention:   0, // Limits policy - allows multiple consumers
    MaxMsgs:     10000,
}
err := h.CreateOrUpdatePersistent(config)

// Multiple workers subscribe to same subject (each with different durable name)
workerNames := []string{"worker-1", "worker-2", "worker-3"}
for _, workerName := range workerNames {
    cancel, err := h.SubscribePersistentViaDurable(workerName, "work.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
        fmt.Printf("Worker %s processing: %s\n", workerName, string(msg))
        return nil, false, true // ACK
    }, func(err error) {
        log.Printf("Worker %s error: %v", workerName, err)
    })
    defer cancel()
}

// Publish task - only one worker will process it
err = h.PublishPersistent("work.tasks", []byte("Important batch job"))
```

**Use cases**: Batch processing, email sending, file conversion

#### WorkQueue Policy vs Limits Policy Comparison

```go
// ❌ WorkQueue policy (only one consumer allowed)
config := &hub.PersistentConfig{
    Subjects:  []string{"work.>"},
    Retention: 2, // WorkQueue - same durable name not allowed
}

// ✅ Limits policy (multiple consumers allowed)
config := &hub.PersistentConfig{
    Subjects:  []string{"work.>"},
    Retention: 0, // Limits - multiple durable names allowed
}
```

| Policy | Consumer Count | Message Processing | Use Cases |
|--------|----------------|-------------------|-----------|
| **WorkQueue** | Only 1 durable name | Exactly once | Strict deduplication |
| **Limits** | Multiple durable names | Load balancing | Task distribution, scalability |

### 5.1 Advanced Persistent QueueSub Patterns

#### Priority Queue
```go
// Create streams for different priorities
priorities := []string{"high", "medium", "low"}
for _, priority := range priorities {
    config := &hub.PersistentConfig{
        Description: fmt.Sprintf("%s priority task queue", priority),
        Subjects:    []string{fmt.Sprintf("work.%s.>", priority)},
        Retention:   0,
        MaxMsgs:     1000,
    }
    err := h.CreateOrUpdatePersistent(config)
    if err != nil {
        log.Printf("Failed to create %s priority queue: %v", priority, err)
    }
}

// Deploy workers for different priorities
h.SubscribePersistentViaDurable("high-priority-worker", "work.high.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("🚨 High priority task: %s\n", string(msg))
    return nil, false, true
}, errHandler)

h.SubscribePersistentViaDurable("medium-priority-worker", "work.medium.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("⚡ Medium priority task: %s\n", string(msg))
    return nil, false, true
}, errHandler)

h.SubscribePersistentViaDurable("low-priority-worker", "work.low.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("🐌 Low priority task: %s\n", string(msg))
    return nil, false, true
}, errHandler)

// Publish tasks with different priorities
h.PublishPersistent("work.high.tasks", []byte("Urgent system check"))
h.PublishPersistent("work.medium.tasks", []byte("User data sync"))
h.PublishPersistent("work.low.tasks", []byte("Log cleanup"))
```

#### Batch Processing Queue
```go
// Create stream for batch jobs
config := &hub.PersistentConfig{
    Description: "Batch processing queue",
    Subjects:    []string{"batch.>"},
    Retention:   0,
    MaxMsgs:     10000,
    MaxAge:      24 * time.Hour,
}
err := h.CreateOrUpdatePersistent(config)

// Batch worker (processes multiple messages at once)
cancel, err := h.SubscribePersistentViaDurable("batch-worker", "batch.jobs", func(subject string, msg []byte) ([]byte, bool, bool) {
    batchID := string(msg)
    fmt.Printf("Starting batch job: %s\n", batchID)
    
    // Process batch
    results := processBatch(batchID)
    
    // Store results in KV store
    resultKey := fmt.Sprintf("batch_result_%s", batchID)
    h.PutToKeyValueStore("batch-results", resultKey, results)
    
    fmt.Printf("Completed batch job: %s\n", batchID)
    return nil, false, true
}, errHandler)

// Publish batch jobs
h.PublishPersistent("batch.jobs", []byte("user_data_sync_2024_01"))
h.PublishPersistent("batch.jobs", []byte("email_campaign_jan"))
```

#### Dead Letter Queue
```go
// Main task queue
mainConfig := &hub.PersistentConfig{
    Description: "Main task queue",
    Subjects:    []string{"work.main.>"},
    Retention:   0,
    MaxMsgs:     1000,
}
err := h.CreateOrUpdatePersistent(mainConfig)

// Dead letter queue (for failed tasks)
dlqConfig := &hub.PersistentConfig{
    Description: "Dead letter queue",
    Subjects:    []string{"work.failed.>"},
    Retention:   0,
    MaxMsgs:     5000,
    MaxAge:      7 * 24 * time.Hour, // 7 days retention
}
err = h.CreateOrUpdatePersistent(dlqConfig)

// Main worker
h.SubscribePersistentViaDurable("main-worker", "work.main.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
    taskID := string(msg)
    success := processTask(taskID)
    
    if !success {
        // Move to dead letter queue on failure
        fmt.Printf("Task failed, moving to DLQ: %s\n", taskID)
        h.PublishPersistent("work.failed.tasks", msg)
        return []byte("moved to DLQ"), false, true
    }
    
    fmt.Printf("Task successful: %s\n", taskID)
    return nil, false, true
}, errHandler)

// Dead letter queue monitor and retry
h.SubscribePersistentViaDurable("dlq-monitor", "work.failed.tasks", func(subject string, msg []byte) ([]byte, bool, bool) {
    failedTaskID := string(msg)
    log.Printf("Failed task detected: %s", failedTaskID)
    
    // Retry logic or admin notification
    notifyAdmin(failedTaskID)
    
    return nil, false, true
}, errHandler)

// Publish main tasks
h.PublishPersistent("work.main.tasks", []byte("payment_process_123"))
h.PublishPersistent("work.main.tasks", []byte("email_send_456"))
```

#### Worker Pool Scaling
```go
// Dynamic worker pool management
type WorkerPool struct {
    hub         *hub.Hub
    subject     string
    workerCount int
    cancels     []func()
}

func (wp *WorkerPool) ScaleTo(count int) error {
    // Clean up existing workers
    for _, cancel := range wp.cancels {
        cancel()
    }
    wp.cancels = nil
    
    // Create new workers
    for i := 0; i < count; i++ {
        workerID := fmt.Sprintf("worker-%d", i)
        cancel, err := wp.hub.SubscribePersistentViaDurable(workerID, wp.subject, func(subject string, msg []byte) ([]byte, bool, bool) {
            fmt.Printf("Worker %s processing: %s\n", workerID, string(msg))
            return nil, false, true
        }, func(err error) {
            log.Printf("Worker %s error: %v", workerID, err)
        })
        
        if err != nil {
            return err
        }
        
        wp.cancels = append(wp.cancels, cancel)
    }
    
    wp.workerCount = count
    return nil
}

// Usage example
pool := &WorkerPool{hub: h, subject: "work.tasks"}
pool.ScaleTo(3)  // Start with 3 workers
// ... Scale based on workload
pool.ScaleTo(10) // Scale to 10 workers
```

#### Message Grouping and Sequential Processing
```go
// Sequential processing queue per user
userConfig := &hub.PersistentConfig{
    Description: "Sequential task queue per user",
    Subjects:    []string{"user.>"}, // user.{userID}.tasks
    Retention:   0,
    MaxMsgs:     1000,
}
err := h.CreateOrUpdatePersistent(userConfig)

// Dedicated worker for each user (same durable name ensures sequential processing)
userIDs := []string{"user123", "user456", "user789"}
for _, userID := range userIDs {
    subject := fmt.Sprintf("user.%s.tasks", userID)
    workerID := fmt.Sprintf("user-worker-%s", userID)
    
    h.SubscribePersistentViaDurable(workerID, subject, func(subject string, msg []byte) ([]byte, bool, bool) {
        fmt.Printf("Processing user %s task: %s\n", userID, string(msg))
        return nil, false, true
    }, errHandler)
}

// Publish tasks per user (each user's tasks are processed sequentially)
h.PublishPersistent("user.user123.tasks", []byte("Profile update"))
h.PublishPersistent("user.user123.tasks", []byte("Password change"))
h.PublishPersistent("user.user456.tasks", []byte("Order cancellation"))
```

### 5.2 Persistent QueueSub Best Practices

#### 1. ACK Strategies
```go
// Immediate ACK (fast processing, no reprocessing)
func fastProcessor(subject string, msg []byte) ([]byte, bool, bool) {
    processQuickly(msg)
    return nil, false, true // Immediate ACK
}

// Delayed ACK (safe processing, reprocessing possible)
func safeProcessor(subject string, msg []byte) ([]byte, bool, bool) {
    if err := processSafely(msg); err != nil {
        return nil, false, false // Reprocess on failure
    }
    return nil, false, true // ACK on success
}

// Conditional ACK
func conditionalProcessor(subject string, msg []byte) ([]byte, bool, bool) {
    result := processWithResult(msg)
    if result.ShouldRetry {
        return nil, false, false // Retry
    }
    return result.Data, false, true // Success
}
```

#### 2. Error Handling and Monitoring
```go
// Worker health check
type WorkerHealth struct {
    lastHeartbeat time.Time
    processedCount int64
    errorCount int64
}

func monitorWorkerHealth(h *hub.Hub) {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()
    
    for range ticker.C {
        // Check worker status and restart logic
        checkWorkerHealth(h)
    }
}

// Metrics collection
func collectMetrics(h *hub.Hub) {
    // Monitor throughput, latency, error rate
    metrics := getQueueMetrics(h)
    reportMetrics(metrics)
}
```

#### 3. Backpressure Handling
```go
// Queue depth monitoring
func monitorQueueDepth(h *hub.Hub, subject string) {
    // Add workers or throttle when queue is full
    depth := getQueueDepth(h, subject)
    
    if depth > 1000 {
        // Add more workers
        addMoreWorkers(h, subject)
    } else if depth < 100 {
        // Remove workers
        removeWorkers(h, subject)
    }
}

// Rate limiting
func rateLimitedProcessor(subject string, msg []byte) ([]byte, bool, bool) {
    if !rateLimiter.Allow() {
        // Wait briefly and reprocess when rate limit exceeded
        time.Sleep(100 * time.Millisecond)
        return nil, false, false
    }
    
    processMessage(msg)
    return nil, false, true
}
```

#### Ephemeral Consumer (Temporary)
```go
// Consumer disappears on system restart
cancel, err := h.SubscribePersistentViaEphemeral("temp.monitor", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("Temporary monitoring: %s\n", string(msg))
    return nil, false, true
}, func(err error) {
    log.Printf("Temporary consumer error: %v", err)
})
```

#### Durable Consumer (Persistent)
```go
// Consumer persists after system restart
cancel, err := h.SubscribePersistentViaDurable("persistent-monitor", "system.events", func(subject string, msg []byte) ([]byte, bool, bool) {
    fmt.Printf("Persistent monitoring: %s\n", string(msg))
    return nil, false, true
}, func(err error) {
    log.Printf("Persistent consumer error: %v", err)
})
```

**Use Cases**:
- **Ephemeral**: Temporary monitoring, debugging, one-time tasks
- **Durable**: Critical event processing, data pipelines, persistent subscriptions

### 7. Key-Value Store (Configuration and Cache)

```go
// Create user settings store
kvConfig := hub.KeyValueStoreConfig{
    Bucket:       "user-settings",
    Description:  "User personal settings",
    MaxValueSize: hub.NewSizeFromKilobytes(64),
    TTL:          30 * 24 * time.Hour, // 30 days
}
err := h.CreateOrUpdateKeyValueStore(kvConfig)

// Store and retrieve settings
settings := `{"theme": "dark", "language": "ko"}`
_, err = h.PutToKeyValueStore("user-settings", "user123", []byte(settings))

data, revision, err := h.GetFromKeyValueStore("user-settings", "user123")
```

**Use Cases**: User settings, application configuration, cached data

### 8. Object Store (Large Files)

```go
// Create document store
objConfig := hub.ObjectStoreConfig{
    Bucket:      "user-documents",
    Description: "User uploaded documents",
    MaxBytes:    hub.NewSizeFromGigabytes(100),
    TTL:         365 * 24 * time.Hour, // 1 year
}
err := h.CreateObjectStore(objConfig)

// Upload file
fileData := readFile("report.pdf")
metadata := map[string]string{
    "filename":    "Monthly Report.pdf",
    "contentType": "application/pdf",
    "uploadedBy":  "user123",
}
err = h.PutToObjectStore("user-documents", "report-2024-01", fileData, metadata)

// Download file
data, err := h.GetFromObjectStore("user-documents", "report-2024-01")
```

**Use Cases**: File uploads, image storage, document management

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

### Volatile Messaging (Volatile Messaging)
- `SubscribeVolatileViaFanout(subject, handler, errHandler)` - Fanout subscription (delivered to all subscribers)
- `SubscribeVolatileViaQueue(subject, queue, handler, errHandler)` - Queue subscription (delivered to only one subscriber)
- `PublishVolatile(subject, msg)` - Publish message
- `RequestVolatile(subject, msg, timeout)` - Publish request and wait for response

### JetStream (Persistent Messaging)
- `CreateOrUpdatePersistent(config)` - Create/update persistent stream
- `SubscribePersistentViaDurable(id, subject, handler, errHandler)` - Durable consumer subscription (persists after restart)
- `SubscribePersistentViaEphemeral(subject, handler, errHandler)` - Ephemeral consumer subscription (temporary)
- `PublishPersistent(subject, msg)` - Publish message to persistent stream

### Key-Value Store (Key-Value Store)
- `CreateOrUpdateKeyValueStore(config)` - Create/update KV store
- `GetFromKeyValueStore(bucket, key)` - Retrieve value by key
- `PutToKeyValueStore(bucket, key, value)` - Store key-value pair
- `UpdateToKeyValueStore(bucket, key, value, expectedRevision)` - Update with version check
- `DeleteFromKeyValueStore(bucket, key)` - Delete key

### Object Store (Object Store)
- `CreateObjectStore(config)` - Create object store
- `GetFromObjectStore(bucket, key)` - Retrieve object
- `PutToObjectStore(bucket, key, data, metadata)` - Store object (with metadata)
- `DeleteFromObjectStore(bucket, key)` - Delete object

### Handler Function Signatures

#### Volatile Messaging Handler
```go
func(subject string, msg []byte) (response []byte, reply bool)
```

#### JetStream Handler
```go
func(subject string, msg []byte) (response []byte, reply bool, ack bool)
```

#### Error Handler
```go
func(error)
```

### ACK (Acknowledgment) Behavior

In JetStream, ACK indicates message processing completion:
- `ack = true`: Message processed successfully, removed from stream
- `ack = false`: Message processing failed, retransmission possible
- `reply = true`: Send response message
- `reply = false`: No response

## Messaging Pattern Selection Guide

| Pattern | Persistence | ACK | QueueSub | Use Cases |
|---------|-------------|-----|----------|-----------|
| **Simple Publish/Subscribe** | ❌ | ❌ | ❌ | Real-time notifications, logging |
| **QueueSub** | ❌ | ❌ | ✅ | Task distribution, load balancing |
| **Request/Reply** | ❌ | ✅ | ❌ | RPC, service calls |
| **JetStream Publish/Subscribe** | ✅ | ✅ | ❌ | Event sourcing, audit logs |
| **JetStream QueueSub** | ✅ | ✅ | ✅ | Batch processing, reliability-critical tasks |
| **Key-Value Store** | ✅ | ✅ | ❌ | Configuration, cache, metadata |
| **Object Store** | ✅ | ✅ | ❌ | File storage, large data |

### Selection Criteria:

- **Real-time required**: Volatile messaging (Publish/Subscribe, QueueSub, Request/Reply)
- **Data persistence required**: JetStream, KV Store, Object Store
- **Task distribution required**: QueueSub pattern
- **Response required**: Request/Reply
- **Large data**: Object Store
- **Frequent read/write**: Key-Value Store

## Advanced Usage

### Cluster Configuration

```go
// Node 1
opts1 := &hub.Options{
    Name:        "node1",
    ClusterHost: "0.0.0.0",
    ClusterPort: 6222,
    Routes:      []*url.URL{}, // URLs of other nodes
}

// Node 2
opts2 := &hub.Options{
    Name:        "node2", 
    ClusterHost: "0.0.0.0",
    ClusterPort: 6223,
    Routes: []*url.URL{
        {Host: "127.0.0.1:6222"}, // Connect to node 1
    },
}
```

### Monitoring and Management

```go
// Stream information retrieval
// (Not directly supported in current API - NATS CLI recommended)

// Consumer status monitoring
// (Not directly supported in current API - NATS CLI recommended)

// Memory usage monitoring
opts := &hub.Options{
    JetstreamMaxMemory: hub.NewSizeFromMegabytes(512),
    StoreDir:           "./data",
}
```

### Error Handling Patterns

```go
// Retry logic
func publishWithRetry(h *hub.Hub, subject string, msg []byte, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := h.PublishVolatile(subject, msg)
        if err == nil {
            return nil
        }
        time.Sleep(time.Duration(i+1) * time.Second)
    }
    return fmt.Errorf("failed after %d retries", maxRetries)
}

// Circuit breaker pattern
type CircuitBreaker struct {
    failureCount int
    lastFailure  time.Time
    state        string // "closed", "open", "half-open"
}

func (cb *CircuitBreaker) Call(fn func() error) error {
    if cb.state == "open" {
        if time.Since(cb.lastFailure) > 30*time.Second {
            cb.state = "half-open"
        } else {
            return fmt.Errorf("circuit breaker is open")
        }
    }
    
    err := fn()
    if err != nil {
        cb.failureCount++
        cb.lastFailure = time.Now()
        if cb.failureCount > 5 {
            cb.state = "open"
        }
        return err
    }
    
    cb.failureCount = 0
    cb.state = "closed"
    return nil
}
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
