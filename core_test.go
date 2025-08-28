package hub

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSubscribeVolatileViaFanout(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_fanout_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Test fanout subscription
	received := make(chan []byte, 10)
	var receivedCount int
	var mu sync.Mutex

	cancel, err := hub.SubscribeVolatileViaFanout("test.fanout", func(subject string, msg []byte) ([]byte, bool) {
		if subject != "test.fanout" {
			t.Errorf("Expected subject 'test.fanout', got '%s'", subject)
		}
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return []byte("response"), true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Publish multiple messages
	testMsgs := [][]byte{
		[]byte("message 1"),
		[]byte("message 2"),
		[]byte("message 3"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishVolatile("test.fanout", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	// Wait for all messages
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs); i++ {
		select {
		case receivedMsg := <-received:
			t.Logf("Received message %d: %s", i+1, string(receivedMsg))
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

func TestSubscribeVolatileViaQueue(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_queue_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Test queue subscription
	received := make(chan []byte, 10)
	var receivedCount int
	var mu sync.Mutex

	cancel, err := hub.SubscribeVolatileViaQueue("test.queue", "worker-group", func(subject string, msg []byte) ([]byte, bool) {
		if subject != "test.queue" {
			t.Errorf("Expected subject 'test.queue', got '%s'", subject)
		}
		mu.Lock()
		receivedCount++
		mu.Unlock()
		received <- msg
		return []byte("processed"), true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Publish messages
	testMsgs := [][]byte{
		[]byte("task 1"),
		[]byte("task 2"),
		[]byte("task 3"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishVolatile("test.queue", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	// Wait for messages (queue group should receive all messages)
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs); i++ {
		select {
		case receivedMsg := <-received:
			t.Logf("Received task %d: %s", i+1, string(receivedMsg))
		case <-timeout:
			t.Fatalf("Timeout waiting for task %d", i+1)
		}
	}

	mu.Lock()
	if receivedCount != len(testMsgs) {
		t.Errorf("Expected %d tasks, got %d", len(testMsgs), receivedCount)
	}
	mu.Unlock()
}

func TestPublishVolatile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_publish_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Test publishing with subscription
	received := make(chan []byte, 1)

	cancel, err := hub.SubscribeVolatileViaFanout("test.publish", func(subject string, msg []byte) ([]byte, bool) {
		received <- msg
		return nil, false // no response
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Test publishing
	testMsg := []byte("test publish message")
	err = hub.PublishVolatile("test.publish", testMsg)
	if err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Wait for message
	select {
	case receivedMsg := <-received:
		if string(receivedMsg) != string(testMsg) {
			t.Errorf("Expected message '%s', got '%s'", string(testMsg), string(receivedMsg))
		}
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for published message")
	}
}

func TestRequestVolatile(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_request_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Set up responder
	cancel, err := hub.SubscribeVolatileViaFanout("test.request", func(subject string, msg []byte) ([]byte, bool) {
		return []byte("response to: " + string(msg)), true
	}, func(err error) {
		t.Errorf("Subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer cancel()

	// Send request
	requestMsg := []byte("hello world")
	response, err := hub.RequestVolatile("test.request", requestMsg, 2*time.Second)
	if err != nil {
		t.Fatalf("Request failed: %v", err)
	}

	expectedResponse := "response to: hello world"
	if string(response) != expectedResponse {
		t.Errorf("Expected response '%s', got '%s'", expectedResponse, string(response))
	}
}

func TestCoreClusterCommunication(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_cluster_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Test multiple subscriptions on same subject (simulating cluster-like behavior)
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)
	var count1, count2 int
	var mu sync.Mutex

	// Subscriber 1 (simulating node 1 in cluster)
	cancel1, err := hub.SubscribeVolatileViaFanout("cluster.broadcast", func(subject string, msg []byte) ([]byte, bool) {
		mu.Lock()
		count1++
		mu.Unlock()
		received1 <- msg
		return []byte("ack from node1"), true
	}, func(err error) {
		t.Errorf("Node 1 subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create node 1 subscriber: %v", err)
	}
	defer cancel1()

	// Subscriber 2 (simulating node 2 in cluster)
	cancel2, err := hub.SubscribeVolatileViaFanout("cluster.broadcast", func(subject string, msg []byte) ([]byte, bool) {
		mu.Lock()
		count2++
		mu.Unlock()
		received2 <- msg
		return []byte("ack from node2"), true
	}, func(err error) {
		t.Errorf("Node 2 subscription error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create node 2 subscriber: %v", err)
	}
	defer cancel2()

	// Publish broadcast messages
	testMsgs := [][]byte{
		[]byte("broadcast message 1"),
		[]byte("broadcast message 2"),
		[]byte("broadcast message 3"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishVolatile("cluster.broadcast", msg)
		if err != nil {
			t.Fatalf("Failed to publish broadcast message: %v", err)
		}
	}

	// Wait for messages on both subscribers
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs)*2; i++ { // 2 subscribers * 3 messages
		select {
		case <-received1:
			t.Logf("Node 1 received broadcast message %d", i+1)
		case <-received2:
			t.Logf("Node 2 received broadcast message %d", i+1)
		case <-timeout:
			t.Fatalf("Timeout waiting for broadcast message %d", i+1)
		}
	}

	mu.Lock()
	if count1 != len(testMsgs) || count2 != len(testMsgs) {
		t.Errorf("Expected each node to receive %d messages, got node1: %d, node2: %d", len(testMsgs), count1, count2)
	}
	mu.Unlock()

	t.Logf("Successfully tested cluster-like broadcast communication")
}

func TestCoreMultipleSubscribers(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_multi_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Create multiple subscribers
	received1 := make(chan []byte, 10)
	received2 := make(chan []byte, 10)
	var count1, count2 int
	var mu sync.Mutex

	// Subscriber 1
	cancel1, err := hub.SubscribeVolatileViaFanout("test.multi", func(subject string, msg []byte) ([]byte, bool) {
		mu.Lock()
		count1++
		mu.Unlock()
		received1 <- msg
		return nil, false
	}, func(err error) {
		t.Errorf("Subscriber 1 error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create subscriber 1: %v", err)
	}
	defer cancel1()

	// Subscriber 2
	cancel2, err := hub.SubscribeVolatileViaFanout("test.multi", func(subject string, msg []byte) ([]byte, bool) {
		mu.Lock()
		count2++
		mu.Unlock()
		received2 <- msg
		return nil, false
	}, func(err error) {
		t.Errorf("Subscriber 2 error: %v", err)
	})

	if err != nil {
		t.Fatalf("Failed to create subscriber 2: %v", err)
	}
	defer cancel2()

	// Publish messages
	testMsgs := [][]byte{
		[]byte("multi 1"),
		[]byte("multi 2"),
		[]byte("multi 3"),
	}

	for _, msg := range testMsgs {
		err = hub.PublishVolatile("test.multi", msg)
		if err != nil {
			t.Fatalf("Failed to publish: %v", err)
		}
	}

	// Wait for messages on both subscribers
	timeout := time.After(5 * time.Second)
	for i := 0; i < len(testMsgs)*2; i++ { // 2 subscribers * 3 messages
		select {
		case <-received1:
			t.Logf("Subscriber 1 received message %d", i+1)
		case <-received2:
			t.Logf("Subscriber 2 received message %d", i+1)
		case <-timeout:
			t.Fatalf("Timeout waiting for message %d", i+1)
		}
	}

	mu.Lock()
	if count1 != len(testMsgs) || count2 != len(testMsgs) {
		t.Errorf("Expected each subscriber to receive %d messages, got sub1: %d, sub2: %d", len(testMsgs), count1, count2)
	}
	mu.Unlock()
}

func TestCoreErrorHandling(t *testing.T) {
	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "core_test_error_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create hub instance
	opts, err := DefaultOptions()
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

	// Test request with no responder (should timeout)
	_, err = hub.RequestVolatile("nonexistent.subject", []byte("test"), 1*time.Second)
	if err == nil {
		t.Error("Expected timeout error for request with no responder")
	} else {
		t.Logf("Got expected error: %v", err)
	}

	// Test publish after shutdown (should fail gracefully)
	hub.Shutdown()

	err = hub.PublishVolatile("test.subject", []byte("test"))
	if err == nil {
		t.Error("Expected error when publishing after shutdown")
	} else {
		t.Logf("Got expected error after shutdown: %v", err)
	}
}
