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
