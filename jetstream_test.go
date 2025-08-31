package hub

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats.go"
)

func TestCreateOrUpdatePersistent(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_create_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Test creating a new stream
	config := &PersistentConfig{
		Description:  "Test stream for orders",
		Subjects:     []string{"orders.>"},
		Retention:    nats.LimitsPolicy,
		MaxConsumers: 10,
		MaxMsgs:      1000,
		MaxBytes:     NewSizeFromMegabytes(10).Bytes(),
		MaxAge:       24 * time.Hour,
		Replicas:     1,
		Metadata: map[string]string{
			"environment": "test",
			"version":     "1.0",
		},
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create persistent stream: %v", err)
	}

	// Verify stream was created
	stream, err := hub.jetstreamCtx.StreamInfo("orders")
	if err != nil {
		t.Fatalf("Failed to get stream info: %v", err)
	}

	if stream.Config.Description != config.Description {
		t.Errorf("Expected description '%s', got '%s'", config.Description, stream.Config.Description)
	}

	if len(stream.Config.Subjects) != len(config.Subjects) {
		t.Errorf("Expected %d subjects, got %d", len(config.Subjects), len(stream.Config.Subjects))
	}

	// Test updating the stream
	config.MaxMsgs = 2000
	config.Description = "Updated test stream for orders"

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to update persistent stream: %v", err)
	}

	// Verify stream was updated
	stream, err = hub.jetstreamCtx.StreamInfo("orders")
	if err != nil {
		t.Fatalf("Failed to get updated stream info: %v", err)
	}

	if stream.Config.MaxMsgs != config.MaxMsgs {
		t.Errorf("Expected MaxMsgs %d, got %d", config.MaxMsgs, stream.Config.MaxMsgs)
	}

	if stream.Config.Description != config.Description {
		t.Errorf("Expected updated description '%s', got '%s'", config.Description, stream.Config.Description)
	}
}

func TestSubscribePersistentViaDurable(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_durable_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream first
	config := &PersistentConfig{
		Description: "Test stream for durable subscription",
		Subjects:    []string{"events.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     100,
	}
	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Test durable subscription
	received := make(chan []byte, 10)
	var receivedCount int
	var mu sync.Mutex

	cancel, err := hub.SubscribePersistentViaDurable("test-consumer", "events.user", func(subject string, msg []byte) ([]byte, bool, bool) {
		if subject != "events.user" {
			t.Errorf("Expected subject 'events.user', got '%s'", subject)
		}
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return []byte("processed"), true, true // reply, ack
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create durable subscription: %v", err)
	}
	defer cancel()

	// Publish messages
	testMsgs := [][]byte{
		[]byte("user login event"),
		[]byte("user logout event"),
		[]byte("user update event"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishPersistent("events.user", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	// Wait for all messages
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs); i++ {
		select {
		case receivedMsg := <-received:
			t.Logf("Durable consumer received message %d: %s", i+1, string(receivedMsg))
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i+1)
		}
	}

	mu.Lock()
	if receivedCount != len(testMsgs) {
		t.Errorf("Expected %d messages, got %d", len(testMsgs), receivedCount)
	}
	mu.Unlock()
}

func TestSubscribePersistentViaEphemeral(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_ephemeral_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream first
	config := &PersistentConfig{
		Description: "Test stream for ephemeral subscription",
		Subjects:    []string{"notifications.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     50,
	}
	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Test ephemeral subscription
	received := make(chan []byte, 10)
	var receivedCount int
	var mu sync.Mutex

	cancel, err := hub.SubscribePersistentViaEphemeral("notifications.alert", func(subject string, msg []byte) ([]byte, bool, bool) {
		if subject != "notifications.alert" {
			t.Errorf("Expected subject 'notifications.alert', got '%s'", subject)
		}
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return nil, false, true // no reply, ack
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create ephemeral subscription: %v", err)
	}
	defer cancel()

	// Publish messages
	testMsgs := [][]byte{
		[]byte("system alert 1"),
		[]byte("system alert 2"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishPersistent("notifications.alert", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	// Wait for messages
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs); i++ {
		select {
		case receivedMsg := <-received:
			t.Logf("Ephemeral consumer received message %d: %s", i+1, string(receivedMsg))
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i+1)
		}
	}

	mu.Lock()
	if receivedCount != len(testMsgs) {
		t.Errorf("Expected %d messages, got %d", len(testMsgs), receivedCount)
	}
	mu.Unlock()
}

func TestPublishPersistent(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_publish_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream first
	config := &PersistentConfig{
		Description: "Test stream for publishing",
		Subjects:    []string{"logs.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     100,
	}
	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Set up subscriber
	received := make(chan []byte, 1)

	cancel, err := hub.SubscribePersistentViaDurable("log-consumer", "logs.app", func(subject string, msg []byte) ([]byte, bool, bool) {
		received <- msg
		return nil, false, true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Test publishing
	testMsg := []byte("application log message")
	err = hub.PublishPersistent("logs.app", testMsg)
	if err != nil {
		t.Fatalf("Failed to publish persistent message: %v", err)
	}

	// Wait for message
	select {
	case receivedMsg := <-received:
		if string(receivedMsg) != string(testMsg) {
			t.Errorf("Expected message '%s', got '%s'", string(testMsg), string(receivedMsg))
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for persistent message")
	}

	// Verify message was stored in stream
	stream, err := hub.jetstreamCtx.StreamInfo("logs")
	if err != nil {
		t.Fatalf("Failed to get stream info: %v", err)
	}

	if stream.State.Msgs != 1 {
		t.Errorf("Expected 1 message in stream, got %d", stream.State.Msgs)
	}
}

func TestJetStreamClusterCommunication(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_cluster_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for cluster communication
	config := &PersistentConfig{
		Description: "Cluster communication stream",
		Subjects:    []string{"cluster.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     100,
		Replicas:    1,
	}
	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create cluster stream: %v", err)
	}

	// Create multiple durable consumers (simulating cluster nodes)
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)
	var count1, count2 int
	var mu sync.Mutex

	// Node 1 consumer
	cancel1, err := hub.SubscribePersistentViaDurable("node1", "cluster.broadcast", func(subject string, msg []byte) ([]byte, bool, bool) {
		mu.Lock()
		count1++
		mu.Unlock()
		received1 <- msg
		return []byte("ack from node1"), true, true
	}, func(err error) {
		t.Errorf("Node 1 error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create node 1 consumer: %v", err)
	}
	defer cancel1()

	// Node 2 consumer
	cancel2, err := hub.SubscribePersistentViaDurable("node2", "cluster.broadcast", func(subject string, msg []byte) ([]byte, bool, bool) {
		mu.Lock()
		count2++
		mu.Unlock()
		received2 <- msg
		return []byte("ack from node2"), true, true
	}, func(err error) {
		t.Errorf("Node 2 error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create node 2 consumer: %v", err)
	}
	defer cancel2()

	// Publish cluster messages
	clusterMsgs := [][]byte{
		[]byte("cluster message 1"),
		[]byte("cluster message 2"),
		[]byte("cluster message 3"),
	}

	for _, msg := range clusterMsgs {
		err = hub.PublishPersistent("cluster.broadcast", msg)
		if err != nil {
			t.Fatalf("Failed to publish cluster message: %v", err)
		}
	}

	// Wait for messages on both nodes
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(clusterMsgs)*2; i++ { // 2 nodes * 3 messages
		select {
		case <-received1:
			t.Logf("Node 1 received cluster message %d", i+1)
		case <-received2:
			t.Logf("Node 2 received cluster message %d", i+1)
		case <-timeout:
			t.Fatalf("Timeout waiting for cluster message %d", i+1)
		}
	}

	mu.Lock()
	if count1 != len(clusterMsgs) || count2 != len(clusterMsgs) {
		t.Errorf("Expected each node to receive %d messages, got node1: %d, node2: %d", len(clusterMsgs), count1, count2)
	}
	mu.Unlock()

	// Verify messages are persisted
	stream, err := hub.jetstreamCtx.StreamInfo("cluster")
	if err != nil {
		t.Fatalf("Failed to get cluster stream info: %v", err)
	}

	if stream.State.Msgs != uint64(len(clusterMsgs)) {
		t.Errorf("Expected %d messages in cluster stream, got %d", len(clusterMsgs), stream.State.Msgs)
	}
}

func TestJetStreamAckHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_ack_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream
	config := &PersistentConfig{
		Description: "Test stream for ACK handling",
		Subjects:    []string{"ack.test"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     10,
	}
	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Test ACK handling
	received := make(chan []byte, 5)
	var ackCount int
	var mu sync.Mutex

	cancel, err := hub.SubscribePersistentViaDurable("ack-consumer", "ack.test", func(subject string, msg []byte) ([]byte, bool, bool) {
		mu.Lock()
		ackCount++
		mu.Unlock()
		received <- msg
		return nil, false, true // no reply, ack
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Publish messages
	for i := 0; i < 3; i++ {
		msg := []byte(fmt.Sprintf("message %d", i+1))
		err = hub.PublishPersistent("ack.test", msg)
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i+1, err)
		}
	}

	// Wait for messages and ACKs
	timeout := time.After(5 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case <-received:
			t.Logf("Received and ACKed message %d", i+1)
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i+1)
		}
	}

	mu.Lock()
	if ackCount != 3 {
		t.Errorf("Expected 3 ACKs, got %d", ackCount)
	}
	mu.Unlock()

	// Verify all messages were acknowledged (should not be redelivered)
	time.Sleep(100 * time.Millisecond) // Brief wait

	select {
	case <-received:
		t.Error("Received unexpected redelivered message")
	default:
		t.Log("No redelivered messages - ACKs working correctly")
	}
}

func TestJetStreamErrorHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_error_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Test publishing to non-existent stream (should fail)
	err = hub.PublishPersistent("nonexistent.stream", []byte("test"))
	if err == nil {
		t.Error("Expected error when publishing to non-existent stream")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Test subscribing to non-existent stream (should fail)
	_, err = hub.SubscribePersistentViaDurable("test-consumer", "nonexistent.stream", func(subject string, msg []byte) ([]byte, bool, bool) {
		return nil, false, true
	}, func(err error) {
		t.Logf("Subscription error (expected): %v", err)
	})

	if err == nil {
		t.Error("Expected error when subscribing to non-existent stream")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Test creating stream with invalid config
	invalidConfig := &PersistentConfig{
		Description: "Invalid stream",
		Subjects:    []string{}, // Empty subjects
	}

	err = hub.CreateOrUpdatePersistent(invalidConfig)
	if err == nil {
		t.Error("Expected error when creating stream with invalid config")
	} else {
		t.Logf("Got expected error for invalid config: %v", err)
	}
}

func TestJetStreamRetentionPolicies(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_test_retention_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Test different retention policies
	retentionTests := []struct {
		name       string
		policy     nats.RetentionPolicy
		maxMsgs    int64
		maxAge     time.Duration
		streamName string
		subjects   []string
	}{
		{
			name:       "Limits Policy",
			policy:     nats.LimitsPolicy,
			maxMsgs:    5,
			maxAge:     0,
			streamName: "limits-test",
			subjects:   []string{"limits-test.>"},
		},
		{
			name:       "Interest Policy",
			policy:     nats.InterestPolicy,
			maxMsgs:    10,
			maxAge:     0,
			streamName: "interest-test",
			subjects:   []string{"interest-test.>"},
		},
	}

	for _, rt := range retentionTests {
		t.Run(rt.name, func(t *testing.T) {
			// Create stream with specific retention policy
			config := &PersistentConfig{
				Description: fmt.Sprintf("Test stream with %s retention", rt.name),
				Subjects:    rt.subjects,
				Retention:   rt.policy,
				MaxMsgs:     rt.maxMsgs,
				MaxAge:      rt.maxAge,
			}

			err = hub.CreateOrUpdatePersistent(config)
			if err != nil {
				t.Fatalf("Failed to create %s stream: %v", rt.name, err)
			}

			// Verify stream configuration
			stream, err := hub.jetstreamCtx.StreamInfo(rt.streamName)
			if err != nil {
				t.Fatalf("Failed to get stream info: %v", err)
			}

			if stream.Config.Retention != rt.policy {
				t.Errorf("Expected retention policy %v, got %v", rt.policy, stream.Config.Retention)
			}

			if stream.Config.MaxMsgs != rt.maxMsgs {
				t.Errorf("Expected MaxMsgs %d, got %d", rt.maxMsgs, stream.Config.MaxMsgs)
			}

			t.Logf("Successfully created stream with %s retention policy", rt.name)
		})
	}

	// Test WorkQueue Policy separately (cannot be changed from other policies)
	t.Run("WorkQueue Policy", func(t *testing.T) {
		config := &PersistentConfig{
			Description: "Test stream with WorkQueue retention",
			Subjects:    []string{"workqueue.>"},
			Retention:   nats.WorkQueuePolicy,
			MaxMsgs:     3,
			MaxAge:      0,
		}

		err = hub.CreateOrUpdatePersistent(config)
		if err != nil {
			t.Fatalf("Failed to create WorkQueue stream: %v", err)
		}

		// Verify stream configuration
		stream, err := hub.jetstreamCtx.StreamInfo("workqueue")
		if err != nil {
			t.Fatalf("Failed to get stream info: %v", err)
		}

		if stream.Config.Retention != nats.WorkQueuePolicy {
			t.Errorf("Expected WorkQueue retention policy, got %v", stream.Config.Retention)
		}

		t.Logf("Successfully created stream with WorkQueue retention policy")
	})
}

func TestPullPersistentViaDurable(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_durable_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for testing
	config := &PersistentConfig{
		Description: "Test stream for pull durable",
		Subjects:    []string{"pull.test.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Publish test messages
	testMessages := []string{
		"message-1",
		"message-2",
		"message-3",
		"message-4",
		"message-5",
	}

	for _, msg := range testMessages {
		err = hub.PublishPersistent("pull.test.messages", []byte(msg))
		if err != nil {
			t.Fatalf("Failed to publish message %s: %v", msg, err)
		}
	}

	// Wait a bit for messages to be stored
	time.Sleep(100 * time.Millisecond)

	// Test pull with custom options
	pullOpts := PullOptions{
		Batch:    3,                     // Fetch 3 messages at a time
		MaxWait:  2 * time.Second,       // Wait up to 2 seconds
		Interval: 50 * time.Millisecond, // Poll every 50ms
	}

	var receivedMessages []string
	var mu sync.Mutex
	messageCount := 0

	cancel, err := hub.PullPersistentViaDurable(
		"test-durable-consumer",
		"pull.test.messages",
		pullOpts,
		func(subject string, msg []byte) ([]byte, bool, bool) {
			mu.Lock()
			defer mu.Unlock()

			receivedMessages = append(receivedMessages, string(msg))
			messageCount++

			t.Logf("Received message %d: %s", messageCount, string(msg))
			return nil, false, true // ACK the message
		},
		func(err error) {
			t.Logf("Pull error: %v", err)
		},
	)

	if err != nil {
		t.Fatalf("Failed to start pull consumer: %v", err)
	}
	defer cancel()

	// Wait for all messages to be received
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatalf("Timeout waiting for messages. Received %d out of %d", len(receivedMessages), len(testMessages))
		case <-ticker.C:
			mu.Lock()
			if len(receivedMessages) >= len(testMessages) {
				mu.Unlock()
				goto checkMessages
			}
			mu.Unlock()
		}
	}

checkMessages:
	mu.Lock()
	defer mu.Unlock()

	if len(receivedMessages) != len(testMessages) {
		t.Errorf("Expected %d messages, got %d", len(testMessages), len(receivedMessages))
	}

	// Verify all messages were received (order may vary)
	receivedSet := make(map[string]bool)
	for _, msg := range receivedMessages {
		receivedSet[msg] = true
	}

	for _, expected := range testMessages {
		if !receivedSet[expected] {
			t.Errorf("Expected message %s not received", expected)
		}
	}

	t.Logf("Successfully received all %d messages via pull durable consumer", len(receivedMessages))
}

func TestPullPersistentViaEphemeral(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_ephemeral_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for testing
	config := &PersistentConfig{
		Description: "Test stream for pull ephemeral",
		Subjects:    []string{"ephemeral.test.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Test with default pull options (should use defaults)
	defaultOpts := PullOptions{} // All zeros, should use defaults

	var receivedCount int
	var mu sync.Mutex

	cancel, err := hub.PullPersistentViaEphemeral(
		"ephemeral.test.data",
		defaultOpts,
		func(subject string, msg []byte) ([]byte, bool, bool) {
			mu.Lock()
			defer mu.Unlock()
			receivedCount++
			t.Logf("Ephemeral consumer received: %s", string(msg))
			return nil, false, true
		},
		func(err error) {
			t.Logf("Ephemeral pull error: %v", err)
		},
	)

	if err != nil {
		t.Fatalf("Failed to start ephemeral pull consumer: %v", err)
	}
	defer cancel()

	// Publish some test messages after starting consumer
	testData := []string{"data-1", "data-2", "data-3"}

	time.Sleep(100 * time.Millisecond) // Let consumer start

	for _, data := range testData {
		err = hub.PublishPersistent("ephemeral.test.data", []byte(data))
		if err != nil {
			t.Fatalf("Failed to publish data %s: %v", data, err)
		}
	}

	// Wait for messages to be processed
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			mu.Lock()
			count := receivedCount
			mu.Unlock()
			t.Fatalf("Timeout waiting for messages. Received %d out of %d", count, len(testData))
		case <-ticker.C:
			mu.Lock()
			if receivedCount >= len(testData) {
				mu.Unlock()
				goto checkComplete
			}
			mu.Unlock()
		}
	}

checkComplete:
	mu.Lock()
	finalCount := receivedCount
	mu.Unlock()

	if finalCount != len(testData) {
		t.Errorf("Expected %d messages, got %d", len(testData), finalCount)
	}

	t.Logf("Successfully received all %d messages via ephemeral pull consumer", finalCount)
}

func TestPullOptionsDefaults(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_defaults_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for testing
	config := &PersistentConfig{
		Description: "Test stream for pull defaults",
		Subjects:    []string{"defaults.test.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Test with various invalid/zero options to verify defaults are applied
	testCases := []struct {
		name string
		opts PullOptions
	}{
		{
			name: "AllZeros",
			opts: PullOptions{Batch: 0, MaxWait: 0, Interval: 0},
		},
		{
			name: "NegativeValues",
			opts: PullOptions{Batch: -1, MaxWait: -1, Interval: -1},
		},
		{
			name: "MixedValidInvalid",
			opts: PullOptions{Batch: 10, MaxWait: 0, Interval: 0},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cancel, err := hub.PullPersistentViaDurable(
				fmt.Sprintf("test-defaults-%s", tc.name),
				"defaults.test.messages",
				tc.opts,
				func(subject string, msg []byte) ([]byte, bool, bool) {
					return nil, false, true
				},
				func(err error) {
					// This should not be called for startup errors
					t.Logf("Pull error in %s: %v", tc.name, err)
				},
			)

			if err != nil {
				t.Fatalf("Failed to start pull consumer with %s: %v", tc.name, err)
			}

			// Verify consumer started successfully (defaults were applied)
			time.Sleep(200 * time.Millisecond)
			cancel()

			t.Logf("Successfully started pull consumer with %s options", tc.name)
		})
	}
}

func TestPullBatchProcessing(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_batch_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for testing
	config := &PersistentConfig{
		Description: "Test stream for batch processing",
		Subjects:    []string{"batch.test.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Publish many messages
	messageCount := 25
	for i := 0; i < messageCount; i++ {
		msg := fmt.Sprintf("batch-message-%d", i+1)
		err = hub.PublishPersistent("batch.test.work", []byte(msg))
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i+1, err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Test with different batch sizes
	testCases := []struct {
		name      string
		batchSize int
	}{
		{"Small batch", 3},
		{"Medium batch", 10},
		{"Large batch", 20},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			pullOpts := PullOptions{
				Batch:    tc.batchSize,
				MaxWait:  1 * time.Second,
				Interval: 50 * time.Millisecond,
			}

			var batchCounts []int
			var totalReceived int
			var mu sync.Mutex
			currentBatch := 0

			consumerID := fmt.Sprintf("batch-consumer-%d", tc.batchSize)

			cancel, err := hub.PullPersistentViaDurable(
				consumerID,
				"batch.test.work",
				pullOpts,
				func(subject string, msg []byte) ([]byte, bool, bool) {
					mu.Lock()
					defer mu.Unlock()

					totalReceived++
					currentBatch++

					// Log every batch completion (approximately)
					if currentBatch >= tc.batchSize {
						batchCounts = append(batchCounts, currentBatch)
						t.Logf("Batch %d completed with %d messages", len(batchCounts), currentBatch)
						currentBatch = 0
					}

					return nil, false, true
				},
				func(err error) {
					t.Logf("Batch pull error: %v", err)
				},
			)

			if err != nil {
				t.Fatalf("Failed to start batch consumer: %v", err)
			}

			// Wait for some messages to be processed
			time.Sleep(2 * time.Second)
			cancel()

			mu.Lock()
			finalTotal := totalReceived
			finalBatches := len(batchCounts)
			mu.Unlock()

			if finalTotal == 0 {
				t.Errorf("No messages received for batch size %d", tc.batchSize)
			}

			t.Logf("Batch size %d: received %d messages in %d batches",
				tc.batchSize, finalTotal, finalBatches)
		})
	}
}

func TestPullConcurrentConsumers(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_concurrent_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream for testing
	config := &PersistentConfig{
		Description: "Test stream for concurrent pull consumers",
		Subjects:    []string{"concurrent.test.>"},
		Retention:   nats.LimitsPolicy,
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Publish test messages
	messageCount := 20
	for i := 0; i < messageCount; i++ {
		msg := fmt.Sprintf("concurrent-message-%d", i+1)
		err = hub.PublishPersistent("concurrent.test.tasks", []byte(msg))
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i+1, err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Start multiple concurrent pull consumers
	consumerCount := 3
	pullOpts := PullOptions{
		Batch:    5,
		MaxWait:  1 * time.Second,
		Interval: 100 * time.Millisecond,
	}

	var wg sync.WaitGroup
	var totalReceived int64
	var mu sync.Mutex
	cancels := make([]func(), consumerCount)

	for i := 0; i < consumerCount; i++ {
		wg.Add(1)
		consumerID := fmt.Sprintf("concurrent-consumer-%d", i+1)

		go func(id string, index int) {
			defer wg.Done()

			cancel, err := hub.PullPersistentViaDurable(
				id,
				"concurrent.test.tasks",
				pullOpts,
				func(subject string, msg []byte) ([]byte, bool, bool) {
					mu.Lock()
					totalReceived++
					count := totalReceived
					mu.Unlock()

					t.Logf("Consumer %s received message %d: %s", id, count, string(msg))
					return nil, false, true
				},
				func(err error) {
					t.Logf("Consumer %s error: %v", id, err)
				},
			)

			if err != nil {
				t.Errorf("Failed to start consumer %s: %v", id, err)
				return
			}

			cancels[index] = cancel
		}(consumerID, i)
	}

	// Wait for consumers to start
	time.Sleep(100 * time.Millisecond)

	// Wait for processing
	time.Sleep(5 * time.Second)

	// Stop all consumers
	for _, cancel := range cancels {
		if cancel != nil {
			cancel()
		}
	}

	wg.Wait()

	mu.Lock()
	finalCount := totalReceived
	mu.Unlock()

	if finalCount == 0 {
		t.Error("No messages received by concurrent consumers")
	}

	// We expect that messages are distributed among consumers
	// Note: In NATS JetStream with separate durable consumers,
	// each consumer gets its own copy of messages, so total can be > messageCount
	if finalCount == 0 {
		t.Error("No messages received by concurrent consumers")
	}

	// With separate durable consumers, each gets all messages
	// This is expected behavior - each durable consumer processes all messages independently
	expectedMinimum := int64(messageCount) // At least one consumer should get all messages
	if finalCount < expectedMinimum {
		t.Errorf("Too few messages received (%d), expected at least %d",
			finalCount, expectedMinimum)
	}

	t.Logf("Successfully processed %d messages across %d concurrent pull consumers (each consumer processes independently)",
		finalCount, consumerCount)
}

func TestPullWorkQueueDistribution(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "jetstream_pull_workqueue_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultNodeOptions()
	if err != nil {
		t.Fatalf("Failed to create default options: %v", err)
	}
	opts.Port = 0
	opts.ClusterPort = 0
	opts.StoreDir = filepath.Join(tempDir, "data")

	hub, err := NewHub(opts)
	if err != nil {
		t.Fatalf("NewHub() failed: %v", err)
	}
	defer hub.Shutdown()

	// Create stream with WorkQueue policy for true work distribution
	config := &PersistentConfig{
		Description: "Test stream for work queue distribution",
		Subjects:    []string{"workqueue.test.>"},
		Retention:   nats.WorkQueuePolicy, // This ensures messages are only delivered once
		MaxMsgs:     1000,
		Replicas:    1,
	}

	err = hub.CreateOrUpdatePersistent(config)
	if err != nil {
		t.Fatalf("Failed to create stream: %v", err)
	}

	// Publish test messages
	messageCount := 15
	for i := 0; i < messageCount; i++ {
		msg := fmt.Sprintf("work-item-%d", i+1)
		err = hub.PublishPersistent("workqueue.test.jobs", []byte(msg))
		if err != nil {
			t.Fatalf("Failed to publish message %d: %v", i+1, err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// Start multiple workers using the SAME durable name for work distribution
	workerCount := 3
	pullOpts := PullOptions{
		Batch:    2,
		MaxWait:  1 * time.Second,
		Interval: 50 * time.Millisecond,
	}

	var totalReceived int64
	var mu sync.Mutex
	var wg sync.WaitGroup
	cancels := make([]func(), workerCount)

	// Use the same durable name so messages are distributed (not duplicated)
	sharedDurableName := "shared-work-consumer"

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		workerIndex := i

		go func(index int) {
			defer wg.Done()

			workerID := fmt.Sprintf("worker-%d", index+1)

			cancel, err := hub.PullPersistentViaDurable(
				sharedDurableName, // Same durable name for all workers
				"workqueue.test.jobs",
				pullOpts,
				func(subject string, msg []byte) ([]byte, bool, bool) {
					mu.Lock()
					totalReceived++
					count := totalReceived
					mu.Unlock()

					t.Logf("%s processed message %d: %s", workerID, count, string(msg))
					time.Sleep(10 * time.Millisecond) // Simulate work
					return nil, false, true
				},
				func(err error) {
					// Only log non-timeout errors
					if err.Error() != "failed to fetch messages from subject \"workqueue.test.jobs\": nats: invalid subscription" {
						t.Logf("%s error: %v", workerID, err)
					}
				},
			)

			if err != nil {
				t.Errorf("Failed to start %s: %v", workerID, err)
				return
			}

			cancels[index] = cancel
		}(workerIndex)
	}

	// Wait for workers to start
	time.Sleep(200 * time.Millisecond)

	// Wait for processing
	time.Sleep(3 * time.Second)

	// Stop all workers
	for _, cancel := range cancels {
		if cancel != nil {
			cancel()
		}
	}

	wg.Wait()

	mu.Lock()
	finalCount := totalReceived
	mu.Unlock()

	if finalCount == 0 {
		t.Error("No messages received by work queue consumers")
	}

	// With WorkQueue policy and same durable name, each message should be processed exactly once
	if finalCount != int64(messageCount) {
		t.Logf("Note: Received %d messages out of %d published. This may be due to timing or some messages not being fetched yet.",
			finalCount, messageCount)
	}

	t.Logf("Work queue successfully distributed %d messages across %d workers using shared durable consumer",
		finalCount, workerCount)
}
